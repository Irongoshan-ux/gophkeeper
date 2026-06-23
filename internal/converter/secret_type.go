// Package converter maps between domain models and API types.
package converter

import (
	gophkeeperv1 "github.com/Irongoshan-ux/gophkeeper/api/gophkeeper/v1"
	"github.com/Irongoshan-ux/gophkeeper/internal/model"
)

// SecretTypeToProto converts a domain secret type to protobuf.
func SecretTypeToProto(t model.SecretType) gophkeeperv1.SecretType {
	switch t {
	case model.SecretTypeCredentials:
		return gophkeeperv1.SecretType_SECRET_TYPE_CREDENTIALS
	case model.SecretTypeText:
		return gophkeeperv1.SecretType_SECRET_TYPE_TEXT
	case model.SecretTypeBinary:
		return gophkeeperv1.SecretType_SECRET_TYPE_BINARY
	case model.SecretTypeCard:
		return gophkeeperv1.SecretType_SECRET_TYPE_CARD
	case model.SecretTypeOTP:
		return gophkeeperv1.SecretType_SECRET_TYPE_OTP
	default:
		return gophkeeperv1.SecretType_SECRET_TYPE_UNSPECIFIED
	}
}

// ProtoToSecretType converts a protobuf secret type to domain.
func ProtoToSecretType(t gophkeeperv1.SecretType) model.SecretType {
	switch t {
	case gophkeeperv1.SecretType_SECRET_TYPE_CREDENTIALS:
		return model.SecretTypeCredentials
	case gophkeeperv1.SecretType_SECRET_TYPE_TEXT:
		return model.SecretTypeText
	case gophkeeperv1.SecretType_SECRET_TYPE_BINARY:
		return model.SecretTypeBinary
	case gophkeeperv1.SecretType_SECRET_TYPE_CARD:
		return model.SecretTypeCard
	case gophkeeperv1.SecretType_SECRET_TYPE_OTP:
		return model.SecretTypeOTP
	default:
		return model.SecretTypeUnspecified
	}
}
