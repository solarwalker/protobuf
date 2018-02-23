package service_register

import (
	pb "github.com/golang/protobuf/protoc-gen-go/descriptor"
	"github.com/golang/protobuf/protoc-gen-go/generator"
)

func init() {
	generator.RegisterPlugin(new(ServiceRegister))
}

type ServiceRegister struct {
	gen *generator.Generator
}

func (sr *ServiceRegister) Name() string {
	return "service_register"
}

func (sr *ServiceRegister) Init(g *generator.Generator) {
	sr.gen = g
}

func (sr *ServiceRegister) Generate(file *generator.FileDescriptor) {
	for i, service := range file.FileDescriptorProto.Service {
		sr.generateService(file, service, i)
	}
}

func (sr *ServiceRegister) GenerateImports(file *generator.FileDescriptor) {
	if len(file.FileDescriptorProto.Service) == 0 {
		return
	}
	sr.gen.P("import (")
	sr.gen.P(`"net/http"`)
	sr.gen.P(")")
	sr.gen.P()
}

func (sr *ServiceRegister) generateService(file *generator.FileDescriptor, service *pb.ServiceDescriptorProto, index int) {
	origServName := service.GetName()
	fullServName := origServName
	if pkg := file.GetPackage(); pkg != "" {
		fullServName = pkg + "." + fullServName
	}

	servName := generator.CamelCase(origServName)
	serverType := servName + "Server"

	sr.gen.P(`// grpc`)
	sr.gen.P(`type GrpcServicer interface {`)
	sr.gen.P(serverType)
	sr.gen.P(`GetServer() *grpc.Server`)
	sr.gen.P(`}`)

	sr.gen.P(`func RegisterGrpcService(s GrpcServicer) {`)
	sr.gen.P(`Register` + serverType + `(s.GetServer(), s)`)
	sr.gen.P(`}`)

	sr.gen.P(`// http`)
	sr.gen.P(`type HttpServicer interface {`)
	sr.gen.P(serverType)
	sr.gen.P(`Handle(pattern string, h http.HandlerFunc)`)
	sr.gen.P(`Decode(ctx context.Context, r *http.Request, arg interface{}) error`)
	sr.gen.P(`HandleReply(ctx context.Context, reply interface{}, w http.ResponseWriter)`)
	sr.gen.P(`}`)

	sr.gen.P(`func RegisterHttpService(s HttpServicer) {`)
	for _, m := range service.Method {
		sr.gen.P(`s.Handle("/api/` + fullServName + `/` + m.GetName() + `", func(writer http.ResponseWriter, request *http.Request) {`)

		sr.gen.P(`ctx := context.Background()`)
		sr.gen.P(`arg := &` + sr.typeName(m.GetInputType()) + `{}`)

		sr.gen.P(`if err := s.Decode(ctx, request, arg); err != nil {`)
		sr.gen.P(`http.Error(writer, err.Error(), http.StatusBadRequest)`)
		sr.gen.P(`return`)
		sr.gen.P(`}`)

		sr.gen.P(`reply, err := s.` + m.GetName() + `(ctx, arg)`)
		sr.gen.P(`if err != nil {`)
		sr.gen.P(`http.Error(writer, err.Error(), http.StatusInternalServerError)`)
		sr.gen.P(`return`)
		sr.gen.P(`}`)
		sr.gen.P(`s.HandleReply(ctx, reply, writer)`)

		sr.gen.P(`})`)
	}
	sr.gen.P(`}`)
}

func (sr *ServiceRegister) typeName(str string) string {
	return sr.gen.TypeName(sr.objectNamed(str))
}

func (sr *ServiceRegister) objectNamed(name string) generator.Object {
	sr.gen.RecordTypeUse(name)
	return sr.gen.ObjectNamed(name)
}
