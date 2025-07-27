// Package errors provides standardized error handling for the Algonius Wallet.
package errors

import (
	"fmt"
)

// ErrorCode represents a specific error type with a unique code
type ErrorCode string

// Error categories
const (
	// Validation Errors
	ErrInvalidParameter     ErrorCode = "INVALID_PARAMETER"
	ErrMissingRequiredField ErrorCode = "MISSING_REQUIRED_FIELD"
	
	// Network Errors
	ErrNetworkConnection ErrorCode = "NETWORK_CONNECTION_ERROR"
	ErrNetworkTimeout    ErrorCode = "NETWORK_TIMEOUT"
	ErrRPCFailure        ErrorCode = "RPC_FAILURE"
	
	// Wallet Errors
	ErrInsufficientBalance ErrorCode = "INSUFFICIENT_BALANCE"
	ErrInvalidAddress      ErrorCode = "INVALID_ADDRESS"
	ErrWalletNotFound      ErrorCode = "WALLET_NOT_FOUND"
	
	// Token Errors
	ErrTokenNotSupported   ErrorCode = "TOKEN_NOT_SUPPORTED"
	ErrInvalidTokenAddress ErrorCode = "INVALID_TOKEN_ADDRESS"
	
	// Permission Errors
	ErrUnauthorized ErrorCode = "UNAUTHORIZED"
	
	// General Errors
	ErrInternal ErrorCode = "INTERNAL_ERROR"
)

// Error represents a standardized error with code, message, details and suggestion
type Error struct {
	Code       ErrorCode `json:"code"`
	Message    string    `json:"message"`
	Details    string    `json:"details,omitempty"`
	Suggestion string    `json:"suggestion,omitempty"`
}

// Error implements the error interface
func (e *Error) Error() string {
	return e.Message
}

// New creates a new Error with the specified code and message
func New(code ErrorCode, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

// WithDetails adds details to the error
func (e *Error) WithDetails(details string) *Error {
	e.Details = details
	return e
}

// WithSuggestion adds a suggestion to the error
func (e *Error) WithSuggestion(suggestion string) *Error {
	e.Suggestion = suggestion
	return e
}

// Wrap wraps an existing error with additional context
func Wrap(err error, code ErrorCode, message string) *Error {
	return &Error{
		Code:    code,
		Message: fmt.Sprintf("%s: %v", message, err),
		Details: err.Error(),
	}
}

// ValidationError creates a validation error
func ValidationError(field, reason string) *Error {
	return New(ErrInvalidParameter, fmt.Sprintf("Invalid '%s' parameter", field)).
		WithDetails(reason).
		WithSuggestion(fmt.Sprintf("Provide a valid value for '%s'", field))
}

// MissingRequiredFieldError creates a missing required field error
func MissingRequiredFieldError(field string) *Error {
	return New(ErrMissingRequiredField, fmt.Sprintf("Missing required field '%s'", field)).
		WithSuggestion(fmt.Sprintf("Include the '%s' parameter in your request", field))
}

// NetworkError creates a network error
func NetworkError(operation string, err error) *Error {
	return Wrap(err, ErrNetworkConnection, fmt.Sprintf("Failed to connect during %s", operation)).
		WithSuggestion("Check your internet connection and try again")
}

// TimeoutError creates a timeout error
func TimeoutError(operation string) *Error {
	return New(ErrNetworkTimeout, fmt.Sprintf("Timeout during %s", operation)).
		WithSuggestion("Try again or check network connectivity")
}

// RPCError creates an RPC error
func RPCError(method string, err error) *Error {
	return Wrap(err, ErrRPCFailure, fmt.Sprintf("RPC call failed for method '%s'", method)).
		WithSuggestion("Check RPC endpoint availability or try again later")
}

// InsufficientBalanceError creates an insufficient balance error
func InsufficientBalanceError(token, balance, required string) *Error {
	return New(ErrInsufficientBalance, fmt.Sprintf("Insufficient balance for token '%s'", token)).
		WithDetails(fmt.Sprintf("Current balance: %s, Required: %s", balance, required)).
		WithSuggestion("Add more funds to your wallet or use a different token")
}

// InvalidAddressError creates an invalid address error
func InvalidAddressError(address, chain string) *Error {
	return New(ErrInvalidAddress, fmt.Sprintf("Invalid address '%s' for chain '%s'", address, chain)).
		WithSuggestion("Check the address format and ensure it's valid for the specified chain")
}

// WalletNotFoundError creates a wallet not found error
func WalletNotFoundError(address string) *Error {
	return New(ErrWalletNotFound, fmt.Sprintf("Wallet not found for address '%s'", address)).
		WithSuggestion("Ensure the wallet exists and is properly imported")
}

// TokenNotSupportedError creates a token not supported error
func TokenNotSupportedError(token, chain string) *Error {
	return New(ErrTokenNotSupported, fmt.Sprintf("Token '%s' is not supported on chain '%s'", token, chain)).
		WithSuggestion(fmt.Sprintf("Use a supported token for chain '%s' or switch to a different chain", chain))
}

// InvalidTokenAddressError creates an invalid token address error
func InvalidTokenAddressError(tokenAddress, chain string) *Error {
	return New(ErrInvalidTokenAddress, fmt.Sprintf("Invalid token address '%s' for chain '%s'", tokenAddress, chain)).
		WithSuggestion("Check the token contract address and ensure it's valid for the specified chain")
}

// UnauthorizedError creates an unauthorized error
func UnauthorizedError(operation string) *Error {
	return New(ErrUnauthorized, fmt.Sprintf("Unauthorized to perform operation '%s'", operation)).
		WithSuggestion("Ensure you have the necessary permissions or authentication")
}

// InternalError creates an internal error
func InternalError(operation string, err error) *Error {
	return Wrap(err, ErrInternal, fmt.Sprintf("Internal error during %s", operation)).
		WithSuggestion("Contact support with error details if the problem persists")
}