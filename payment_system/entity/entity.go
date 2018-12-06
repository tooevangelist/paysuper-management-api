package entity

const (
	BankCardFieldPan       = "pan"
	BankCardFieldCvv       = "cvv"
	BankCardFieldMonth     = "month"
	BankCardFieldYear      = "year"
	BankCardFieldHolder    = "card_holder"
	EWalletFieldIdentifier = "ewallet"
	CryptoFieldIdentifier  = "address"

	TxnParamsFieldBankCardEmissionCountry = "emission_country"
	TxnParamsFieldBankCardToken           = "token"
	TxnParamsFieldBankCardIs3DS           = "is_3ds"
	TxnParamsFieldBankCardRrn             = "rrn"
	TxnParamsFieldDeclineCode             = "decline_code"
	TxnParamsFieldDeclineReason           = "decline_reason"
	TxnParamsFieldCryptoTransactionId     = "transaction_id"
	TxnParamsFieldCryptoAmount            = "amount_crypto"
	TxnParamsFieldCryptoCurrency          = "currency_crypto"
)

var PaymentRequisitesNames = map[string]string{
	BankCardFieldPan: "Card number",
	BankCardFieldHolder: "Card holder",
	EWalletFieldIdentifier: "Payment system account",
	CryptoFieldIdentifier: "Address",
	TxnParamsFieldBankCardIs3DS: "Is 3DS Secured",
	TxnParamsFieldDeclineCode: "Decline code",
	TxnParamsFieldDeclineReason: "Decline reason",
	TxnParamsFieldCryptoAmount: "Amount in crypto currency",
}
