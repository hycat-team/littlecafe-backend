package media

import (
	"context"
	"github.com/hycat-team/littlecafe-backend/shared/application/dto"
)

type IMediaService interface {
	GenerateUploadSignature(ctx context.Context, params dto.UploadSignatureParams) (*dto.UploadSignatureResult, error)
	DeleteImage(ctx context.Context, publicID string) error
}
