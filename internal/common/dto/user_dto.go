package dto

type UserLoginRequestDTO struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type CreateUserRequestDTO struct {
	UserLoginRequestDTO
}

type CreateUserDTO struct {
	Login        string `json:"login"`
	PasswordHash string `json:"password_hash"`
}

type UserDTO struct {
	Login string `json:"login"`
	ID    string `json:"id"`
}

//type AuRequest struct {
//	ID    string            `json:"id"`   // имя метрики.
//	MType common.MetricType `json:"type"` // параметр, принимающий значение gauge или counter.
//}
//
//func parseType(metricTypeAsString string) (common.MetricType, error) {
//	mType, ok := common.StringToMetricType(metricTypeAsString)
//	if !ok {
//		return common.UnknownMetricType, fmt.Errorf(common.ErrUnknownMetric, metricTypeAsString)
//	}
//
//	if mType == common.UnknownMetricType {
//		return common.UnknownMetricType, fmt.Errorf(common.ErrUnknownMetric, metricTypeAsString)
//	}
//
//	return mType, nil
//}
//
//func (gr *GetValueRequest) UnmarshalJSON(data []byte) (err error) {
//	type ReqAlias GetValueRequest
//
//	aliasValue := &struct {
//		*ReqAlias
//		MType string `json:"type"`
//	}{
//		ReqAlias: (*ReqAlias)(gr),
//	}
//
//	if err = json.Unmarshal(data, aliasValue); err != nil {
//		return fmt.Errorf("unmarshal GetValueRequest: %w", err)
//	}
//
//	mType, parseError := parseType(aliasValue.MType)
//	if parseError != nil {
//		return fmt.Errorf("unmarshal GetValueRequest: %w", parseError)
//	}
//
//	gr.MType = mType
//	return nil
//}
//
//type MetricDTO struct {
//	Delta *int64            `json:"delta,omitempty"`
//	Value *float64          `json:"value,omitempty"`
//	ID    string            `json:"id"`
//	MType common.MetricType `json:"type"`
//}
//
//func (m MetricDTO) MarshalJSON() ([]byte, error) {
//	type DtoAlias MetricDTO
//
//	aliasValue := struct {
//		DtoAlias
//		MType string `json:"type"`
//	}{
//		DtoAlias: DtoAlias(m),
//		MType:    m.MType.String(),
//	}
//
//	bytes, err := json.Marshal(aliasValue)
//	if err != nil {
//		return nil, fmt.Errorf("marshal MetricDTO: %w", err)
//	}
//	return bytes, nil
//}
//
//func (m *MetricDTO) UnmarshalJSON(data []byte) (err error) {
//	type DtoAlias MetricDTO
//
//	aliasValue := &struct {
//		*DtoAlias
//		MType string `json:"type"`
//	}{
//		DtoAlias: (*DtoAlias)(m),
//	}
//
//	if err = json.Unmarshal(data, aliasValue); err != nil {
//		return fmt.Errorf("unmarshal MetricDTO: %w", err)
//	}
//
//	mType, parseError := parseType(aliasValue.MType)
//	if parseError != nil {
//		return fmt.Errorf("unmarshal MetricDTO type: %w", parseError)
//	}
//
//	m.MType = mType
//	return nil
//}
