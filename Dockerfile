FROM goreleaser/goreleaser:v1.19.2 AS builder

WORKDIR /go/src/github.com/codekoala/k8s-pvc-usage

RUN apk add --update upx

COPY . .

RUN goreleaser build --snapshot --clean

FROM gcr.io/distroless/static-debian11

COPY --from=builder /go/src/github.com/codekoala/k8s-pvc-usage/dist/k8s-pvc-usage /

CMD ["/k8s-pvc-usage"]

EXPOSE 9100
