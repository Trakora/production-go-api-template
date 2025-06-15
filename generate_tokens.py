import argparse
import secrets
import base64
import sys


def generate_secure_string(num_bytes: int) -> str:
    raw = secrets.token_bytes(num_bytes)
    encoded = base64.urlsafe_b64encode(raw).decode('utf-8')
    return encoded.rstrip('=')


def main():
    parser = argparse.ArgumentParser(
        description="Generate secure API access tokens and secrets"
    )
    parser.add_argument(
        "--secret-length",
        type=int,
        default=64,
        help="Number of random bytes for the secret (default: 64)"
    )
    parser.add_argument(
        "--token-length",
        type=int,
        default=64,
        help="Number of random bytes for the API access token (default: 64)"
    )
    args = parser.parse_args()

    if args.secret_length <= 0 or args.token_length <= 0:
        print("Error: both lengths must be positive integers.", file=sys.stderr)
        sys.exit(1)

    secret = generate_secure_string(args.secret_length)
    token = generate_secure_string(args.token_length)

    print(f"Generated Secure Secret: {secret}")
    print(f"Generated API Access Token: {token}")


if __name__ == "__main__":
    main()