package processor

import (
	"context"
	"errors"

	ausf_sor_protection "github.com/ShouheiNishi/openapi5g/ausf/sor"
	"github.com/ShouheiNishi/openapi5g/utils/error/middleware"
	strictgin "github.com/oapi-codegen/runtime/strictmiddleware/gin"
)

func NewServerAusfSorProtection(processor *Processor) ausf_sor_protection.ServerInterface {
	return ausf_sor_protection.NewStrictHandler(
		&ausfSorProtectionStrictServer{
			processor: processor,
		},
		[]strictgin.StrictGinMiddlewareFunc{middleware.GinStrictServerMiddleware},
	)
}

type ausfSorProtectionStrictServer struct {
	processor *Processor
}

// (POST /{supi}/ue-sor)
func (s *ausfSorProtectionStrictServer) PostSupiUeSor(ctx context.Context,
	request ausf_sor_protection.PostSupiUeSorRequestObject,
) (ausf_sor_protection.PostSupiUeSorResponseObject, error) {
	return nil, errors.New("not implemented")
}