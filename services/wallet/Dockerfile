# Build context MUST be the repository root:
#   docker build -f services/wallet/Dockerfile .
FROM golang:1.26.0 AS builder
WORKDIR /src

COPY go.work ./
COPY shared/ shared/
COPY contracts/ contracts/
COPY services/ services/
COPY configs/ configs/

RUN --mount=type=cache,target=/go/pkg/mod \
	--mount=type=cache,target=/root/.cache/go-build \
	go work sync \
	&& CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" \
		-o /out/app ./services/wallet/cmd/wallet

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app

COPY --from=builder /out/app /app/app
COPY --from=builder /src/configs/ configs/
COPY --from=builder /src/services/wallet/database/migrations/ database/migrations/

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
	CMD ["/app/app", "-healthcheck"]

ENTRYPOINT ["/app/app"]
