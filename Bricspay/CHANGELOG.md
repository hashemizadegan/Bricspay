# CHANGELOG - BRICS Pay Core Package

## [v1.1.0-metamask-eip191] - 2025-05-18
### Added
- **EIP-191 Personal Sign Authentication**: Implementation of cryptographically secure wallet authentication via MetaMask using `github.com/ethereum/go-ethereum/crypto`.
- **Wallet Challenge Lifecycle**: Replay-attack resistant challenge generation (`/api/v1/auth/wallet/challenge`) with dynamic nonce, issued-at, domain, and 5-minute single-use expiration.
- **Verification Endpoint**: `/api/v1/auth/wallet/verify` recovers signer address from Keccak256 hashed EIP-191 personal sign payload, verifies nonce validity, marks challenge consumed, and automatically registers or links member user accounts.
- **PostgreSQL Schemas**: Added `wallet_accounts` and `wallet_challenges` tables with index optimizations and cascade associations.
- **Frontend MetaMask Integration**: Updated `auth.js` and `index.html` with direct `window.ethereum.request({ method: 'personal_sign' })` workflow, dynamic status feedback, and JWT session persistence.
- **Deployment Guidance**: Added comprehensive Railway and Docker environment variables documentation (`DATABASE_URL`, `JWT_SECRET`, `PORT`, `UPLOAD_DIR`).
