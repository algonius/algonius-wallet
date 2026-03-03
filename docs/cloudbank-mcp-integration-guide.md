# CloudBank x Algonius Wallet MCP Integration Guide (KR4)

This guide documents how to integrate `algonius-wallet` MCP tools into CloudBank and how KR4 validation was executed.

## 1. Scope

Target project: `SuLabsOrg/CloudBank`  
Integration objective: run an end-to-end MCP transaction path in CloudBank context:

1. `create_wallet`
2. `fund` (testnet faucet step)
3. `sign_message`
4. `send_transaction`
5. `get_transaction_status`

## 2. MCP Server Configuration

Start native host:

```bash
cd native
make build
SSE_PORT=:9444 SSE_BASE_URL=http://127.0.0.1:9444/mcp ./bin/algonius-wallet-host
```

CloudBank-side MCP endpoint:

- Streamable HTTP: `http://127.0.0.1:9444/mcp`
- SSE: `http://127.0.0.1:9444/mcp/sse`

## 3. CloudBank Faucet Context

CloudBank web implementation references:

- BSC testnet target chain: `97`
- tBNB faucet page: `https://www.bnbchain.org/en/testnet-faucet`
- Example test token addresses used by CloudBank docs:
  - tUSDT: `0x907f4DAA6Ff8083EBdb60FC548603bA79DC970f6`
  - tCOD: `0x6299960264AC6c64592AcAaad96b647d0BaeF1C1`

For automated KR4 validation, faucet funding step supports:

- Real mode: set `CLOUDBANK_FAUCET_API_URL` (if a service endpoint is available)
- CI/mock mode: no env var set, test records a mock funding marker

## 4. Tool Call Examples

### 4.1 Create wallet

```json
{
  "name": "create_wallet",
  "arguments": {
    "chain": "bsc"
  }
}
```

### 4.2 Sign CloudBank payload

```json
{
  "name": "sign_message",
  "arguments": {
    "address": "<wallet_address>",
    "message": "cloudbank-predict-order:<wallet_address>:<timestamp>"
  }
}
```

### 4.3 Send transaction (BSC)

```json
{
  "name": "send_transaction",
  "arguments": {
    "chain": "bsc",
    "from": "<wallet_address>",
    "to": "0x907f4DAA6Ff8083EBdb60FC548603bA79DC970f6",
    "amount": "0.001",
    "token": "BNB"
  }
}
```

### 4.4 Confirm status

```json
{
  "name": "get_transaction_status",
  "arguments": {
    "transaction_hash": "<tx_hash>",
    "chain": "bsc"
  }
}
```

## 5. Error Handling Patterns

Recommended handling for external projects:

- `INVALID_ADDRESS`: reject invalid `from`/`to`, prompt user to re-check chain-specific format.
- `MISSING_REQUIRED_FIELD`: retry after supplying required parameters.
- `NETWORK_TIMEOUT` / `RPC_FAILURE`: retry with backoff and alternate RPC endpoint.
- `INSUFFICIENT_BALANCE`: trigger faucet/deposit flow before retrying.

KR4 reliability check also verifies recovery: an invalid `send_transaction` call is followed by a valid call in the same session and succeeds.

## 6. KR4 Validation Record

Validation test file:

- `native/tests/integration/cloudbank_mcp_integration_test.go`

Commands:

```bash
cd native
make cloudbank-integration-test
```

Observed validation points:

- E2E sequence completed in one MCP session
- Consecutive calls (`sign_message` + `get_balance`) remained stable
- Error recovery path passed (`invalid from` -> next valid send success)

Recorded run (local):

- Date: `2026-03-03`
- Command: `cd native && make cloudbank-integration-test`
- Result: `PASS` (`TestCloudBankMCPIntegration`)
- Sub-steps passed:
  - `create_wallet`
  - `fund_via_cloudbank_testnet_faucet`
  - `sign_transaction_payload`
  - `send_transaction`
  - `confirm_transaction`
  - `stability_consecutive_calls`
  - `error_recovery`

## 7. Known Boundary

Current `send_transaction` tool focuses on transfer flow. For contract-function calls with custom calldata (for example direct faucet contract `claim*` invocation), external project should currently use its own signer/runtime path.
