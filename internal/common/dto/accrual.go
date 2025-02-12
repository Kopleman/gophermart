package dto

type AccrualStatusType string

const (
	AccrualStatusTypeREGISTERED AccrualStatusType = "REGISTERED"
	AccrualStatusTypeINVALID    AccrualStatusType = "INVALID"
	AccrualStatusTypePROCESSING AccrualStatusType = "PROCESSING"
	AccrualStatusTypePROCESSED  AccrualStatusType = "PROCESSED"
)

var (
	accrualStatusTypeName = map[AccrualStatusType]string{
		AccrualStatusTypeREGISTERED: "REGISTERED",
		AccrualStatusTypeINVALID:    "INVALID",
		AccrualStatusTypePROCESSING: "PROCESSING",
		AccrualStatusTypePROCESSED:  "PROCESSED",
	}
	accrualStatusTypeValue = map[string]AccrualStatusType{
		"REGISTERED": AccrualStatusTypeREGISTERED,
		"INVALID":    AccrualStatusTypeINVALID,
		"PROCESSING": AccrualStatusTypePROCESSING,
		"PROCESSED":  AccrualStatusTypePROCESSED,
	}
)

func (ast AccrualStatusType) String() string {
	return accrualStatusTypeName[ast]
}

func StringToAccrualStatusType(s string) (AccrualStatusType, bool) {
	res, ok := accrualStatusTypeValue[s]
	return res, ok
}

type AccrualResponseDTO struct {
	Accrual *float64          `json:"accrual,omitempty"`
	Order   string            `json:"order"`
	Status  AccrualStatusType `json:"status"`
}
