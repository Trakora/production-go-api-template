package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net"
	"net/http"
	"production-go-api-template/config"
	"production-go-api-template/pkg/constants"
	"production-go-api-template/pkg/logger"
	"production-go-api-template/pkg/router"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	splitParts             = 2
	secondPartIndex        = 1
	resetFailuresTo        = 0
	cleanupMultiplier      = 10
	delayThreshold         = 0
	baseDecimal            = 10
	allowedTimeSkewSeconds = 300
)

type ipState struct {
	failures     int
	blockedUntil time.Time
	lastSeen     time.Time
	slowdown     time.Duration
}

type Authenticator struct {
	apiToken      string
	hmacSecret    string
	ipFailures    map[string]*ipState
	ipBlocks      map[string]time.Time
	mu            sync.Mutex
	maxFailures   int
	failWindow    time.Duration
	blockDuration time.Duration
	cleanupTick   time.Duration
	slowdownStep  time.Duration
	slowdownMax   time.Duration
	log           *logger.Logger
}

func NewAuthenticator(authCfg config.ConfAuth, secCfg config.ConfSecurity, log *logger.Logger) *Authenticator {
	a := &Authenticator{
		apiToken:      authCfg.APITokens,
		hmacSecret:    authCfg.HMACSecrets,
		ipFailures:    make(map[string]*ipState),
		ipBlocks:      make(map[string]time.Time),
		maxFailures:   secCfg.MaxFailures,
		failWindow:    secCfg.FailWindow,
		blockDuration: secCfg.BlockDuration,
		cleanupTick:   secCfg.CleanupTick,
		slowdownStep:  secCfg.SlowdownStep,
		slowdownMax:   secCfg.SlowdownMax,
		log:           log,
	}

	go a.cleanupLoop()
	return a
}

func (a *Authenticator) Middleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r)

			if a.handleBlockedIP(r, w, ip) {
				return
			}

			a.handleSlowdown(ip)

			token, ok := a.extractBearerToken(w, r, ip)
			if !ok {
				return
			}

			if token != a.apiToken {
				a.recordFailure(ip)
				router.RespondWithError(r, w, http.StatusForbidden, "invalid token", nil)
				return
			}

			if !a.handleSignature(w, r, ip, token) {
				return
			}

			a.resetIP(ip)

			next.ServeHTTP(w, r)
		})
	}
}

func (a *Authenticator) handleBlockedIP(r *http.Request, w http.ResponseWriter, ip string) bool {
	if a.isBlocked(ip) {
		a.log.Warnf("Blocked request from IP %s to %s %s", ip, r.Method, r.URL.Path)
		router.RespondWithError(r, w, http.StatusForbidden, "blocked", nil)
		return true
	}
	return false
}

func (a *Authenticator) handleSlowdown(ip string) {
	if delay := a.getSlowdown(ip); delay > time.Duration(delayThreshold) {
		time.Sleep(delay)
	}
}

func (a *Authenticator) extractBearerToken(w http.ResponseWriter, r *http.Request, ip string) (string, bool) {
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		a.recordFailure(ip)
		router.RespondWithError(r, w, http.StatusUnauthorized, "missing bearer", nil)
		return "", false
	}
	return strings.TrimPrefix(auth, "Bearer "), true
}

func (a *Authenticator) handleSignature(w http.ResponseWriter, r *http.Request, ip, token string) bool {
	signature := r.Header.Get("X-Signature")
	timestampStr := r.Header.Get("X-Timestamp")

	if signature == constants.EmptyString || timestampStr == constants.EmptyString {
		a.recordFailure(ip)
		a.log.Warnf("Missing signature or timestamp from IP %s", ip)
		router.RespondWithError(r, w, http.StatusUnauthorized, "missing signature or timestamp", nil)
		return false
	}

	ts, err := strconv.ParseInt(timestampStr, baseDecimal, constants.BigInt)
	if err != nil {
		a.recordFailure(ip)
		a.log.Warnf("Invalid timestamp format from IP %s", ip)
		router.RespondWithError(r, w, http.StatusUnauthorized, "invalid timestamp", nil)
		return false
	}
	now := time.Now().Unix()
	if ts > now+allowedTimeSkewSeconds || ts < now-allowedTimeSkewSeconds {
		a.recordFailure(ip)
		a.log.Warnf("Timestamp out of range from IP %s", ip)
		router.RespondWithError(r, w, http.StatusUnauthorized, "timestamp out of range", nil)
		return false
	}

	message := token + constants.PipeSeparator + timestampStr + constants.PipeSeparator + r.Method + constants.PipeSeparator + r.URL.Path
	if !validateHMAC(message, signature, a.hmacSecret) {
		a.recordFailure(ip)
		a.log.Warnf("Invalid HMAC signature from IP %s", ip)
		router.RespondWithError(r, w, http.StatusForbidden, "invalid signature", nil)
		return false
	}

	return true
}

func validateHMAC(message, signature, secret string) bool {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(message))
	expectedMAC := mac.Sum(nil)
	sigBytes, err := hex.DecodeString(signature)
	if err != nil {
		return false
	}
	return hmac.Equal(expectedMAC, sigBytes)
}

func clientIP(r *http.Request) string {
	if h := r.Header.Get("X-Forwarded-For"); h != "" {
		parts := strings.Split(h, ",")
		return strings.TrimSpace(parts[constants.ZeroIndex])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func (a *Authenticator) isBlocked(ipStr string) bool {
	a.mu.Lock()
	defer a.mu.Unlock()

	now := time.Now()

	if state, ok := a.ipFailures[ipStr]; ok && now.Before(state.blockedUntil) {
		return true
	}

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}

	ip4 := ip.To4()
	if ip4 != nil {
		subnet := ip4.Mask(net.CIDRMask(24, 32)).String()
		blockUntil, blocked := a.ipBlocks[subnet]
		return blocked && now.Before(blockUntil)
	}

	return false
}

func (a *Authenticator) recordFailure(ipStr string) {
	a.mu.Lock()
	defer a.mu.Unlock()

	now := time.Now()
	state, exists := a.ipFailures[ipStr]
	if !exists {
		state = &ipState{}
		a.ipFailures[ipStr] = state
	}

	if now.Sub(state.lastSeen) > a.failWindow {
		state.failures = resetFailuresTo
	}
	state.failures++
	state.lastSeen = now

	a.log.Infof("Recording failure for IP %s (total failures: %d/%d)", ipStr, state.failures, a.maxFailures)

	if state.slowdown < a.slowdownMax {
		state.slowdown += a.slowdownStep
		if state.slowdown == a.slowdownMax {
			a.log.Infof("IP %s reached maximum slowdown of %s", ipStr, a.slowdownMax)
		}
	}

	if state.failures >= a.maxFailures {
		state.blockedUntil = now.Add(a.blockDuration)
		a.log.Warnf("Blocking IP %s for %s after %d failures", ipStr, a.blockDuration, state.failures)
		a.blockSubnet(ipStr)
	}
}

func (a *Authenticator) blockSubnet(ipStr string) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return
	}
	if ip.IsLoopback() {
		return
	}
	ip4 := ip.To4()
	if ip4 == nil {
		return
	}
	subnet := ip4.Mask(net.CIDRMask(24, 32)).String()
	a.ipBlocks[subnet] = time.Now().Add(a.blockDuration)
	a.log.Warnf("Blocking subnet %s/24 for %s", subnet, a.blockDuration)
}

func (a *Authenticator) resetIP(ipStr string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if _, exists := a.ipFailures[ipStr]; exists {
		a.log.Infof("Resetting failure count for IP %s after successful authentication", ipStr)
		delete(a.ipFailures, ipStr)
	}
}

func (a *Authenticator) getSlowdown(ipStr string) time.Duration {
	a.mu.Lock()
	defer a.mu.Unlock()
	if state, ok := a.ipFailures[ipStr]; ok {
		return state.slowdown
	}
	return time.Duration(resetFailuresTo)
}

func (a *Authenticator) cleanupLoop() {
	ticker := time.NewTicker(a.cleanupTick)
	defer ticker.Stop()

	for range ticker.C {
		a.mu.Lock()
		now := time.Now()

		ipCountBefore := len(a.ipFailures)
		subnetCountBefore := len(a.ipBlocks)

		for ip, state := range a.ipFailures {
			if now.Sub(state.lastSeen) > a.failWindow*time.Duration(cleanupMultiplier) {
				delete(a.ipFailures, ip)
			}
		}
		for subnet, until := range a.ipBlocks {
			if until.Before(now) {
				delete(a.ipBlocks, subnet)
			}
		}

		ipCountAfter := len(a.ipFailures)
		subnetCountAfter := len(a.ipBlocks)

		a.mu.Unlock()

		a.log.Infof("Security stats: %d active IP blocks, %d active subnet blocks. Cleaned %d IPs, %d subnets.",
			ipCountAfter, subnetCountAfter,
			ipCountBefore-ipCountAfter, subnetCountBefore-subnetCountAfter)
	}
}
