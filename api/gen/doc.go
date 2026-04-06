// Package apigen 包含由 OpenAPI 规范生成的类型、嵌入的 spec 与 Gin 路由注册代码。
// 禁止手改 apigen.gen.go；修改接口请编辑 api/openapi/openapi.yaml 后重新生成。
package apigen

//go:generate go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.6.0 -config ../openapi/oapi-codegen.yaml ../openapi/openapi.yaml
