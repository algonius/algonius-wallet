# Algonius Wallet

Algonius Wallet is a browser extension + Go native host wallet stack for AI agents and users.

## Requirements

- Node.js 20+
- npm 10+
- Go 1.24+
- GNU Make

## Quick Start

1. Install frontend dependencies:

```bash
npm ci
```

2. Build extension (includes TypeScript type check):

```bash
npm run build
```

3. Build native host:

```bash
cd native
make build
```

4. Run native host locally:

```bash
cd native
make run
```

## Test Commands

- Frontend unit tests: `npm run test:ci`
- Native non-integration tests: `cd native && make unit-test`
- Native integration tests: `cd native && make integration-test-all`
- Native transaction-flow e2e tests: `cd native && make e2e-test`
  - Integration tests are behind build tag `integration`
  - Default `go test ./...` will not execute integration tests

## Supported Chains

- `ethereum`
- `bsc`
- `solana`

Source of truth:
- MCP resource `chains://supported`
- Wallet status resource `wallet://status`

## MCP Resources

- `chains://supported`: returns supported chain list
- `wallet://status`: returns current wallet readiness/address/public key/chains

## MCP Tools

- `create_wallet`
- `get_balance`
- `send_transaction`
- `estimate_gas`
- `approve_transaction`
- `swap_tokens`
- `get_pending_transactions`
- `get_transaction_history`
- `deploy_contract`
- `call_contract`
- `simulate_transaction`
- `sign_message`
- `get_transaction_status`

Runtime behavior:
- Chain aliases are normalized (`eth`/`ethereum`, `bsc`/`binance`, `sol`/`solana`)
- Tool RPC paths use timeout + retry wrappers for transient failures
- Errors follow standardized code/message/details/suggestion format

## Architecture Overview

The project has two runtime components:

- Browser extension (`src/`): popup UI, background/content scripts, native messaging bridge
- Native host (`native/`): wallet engine, chain adapters, MCP server, native messaging handlers

Communication model:

1. Extension sends RPC via Chrome Native Messaging to native host.
2. Native host handles secure wallet operations and exposes MCP tools/resources for AI clients.
3. MCP is available over Streamable HTTP and SSE endpoints.

## CI

`/.github/workflows/ci.yml` provides:

- Native host build + lint + non-integration tests
- Extension lint + build + unit tests
- Optional manual runs for integration tests and Playwright e2e tests (`workflow_dispatch`)
- Optional manual run for native MCP e2e flow tests: `run_native_e2e=true`

## Repo Layout

- `src/`: extension source code
- `native/`: Go native host source code and Makefile
- `e2e/`: Playwright e2e tests
- `docs/`: requirements/design/task docs
