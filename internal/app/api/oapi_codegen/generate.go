//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -config ../../../../configs/oapi-codegen.cfg.models.yml ../../../../docs/openapi-models.yml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -config ../../../../configs/oapi-codegen.cfg.public.yml ../../../../docs/openapi-public.yml
//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen -config ../../../../configs/oapi-codegen.cfg.authorized.yml ../../../../docs/openapi-authorized.yml

package oapi_codegen
