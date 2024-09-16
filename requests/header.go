package requests

import (
	"context"
	"encoding/binary"
	"github.com/Chronicle20/atlas-tenant"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"net/http"
)

type HeaderDecorator func(header http.Header)

//goland:noinspection GoUnusedExportedFunction
func SpanHeaderDecorator(ctx context.Context) HeaderDecorator {
	return func(h http.Header) {
		carrier := propagation.MapCarrier{}
		propagator := otel.GetTextMapPropagator()
		propagator.Inject(ctx, carrier)
		for _, k := range carrier.Keys() {
			h.Set(k, carrier.Get(k))
		}
	}
}

//goland:noinspection GoUnusedExportedFunction
func TenantHeaderDecorator(ctx context.Context) HeaderDecorator {
	return func(h http.Header) {
		h.Set("Content-Type", "application/json; charset=utf-8")

		t, err := tenant.FromContext(ctx)()
		if err != nil {
			return
		}

		h.Set(tenant.ID, t.Id().String())
		h.Set(tenant.Region, t.Region())
		h.Set(tenant.MajorVersion, string(binary.BigEndian.AppendUint16(make([]byte, 0), t.MajorVersion())))
		h.Set(tenant.MinorVersion, string(binary.BigEndian.AppendUint16(make([]byte, 0), t.MinorVersion())))
	}
}
