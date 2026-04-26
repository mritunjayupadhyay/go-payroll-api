# syntax=docker/dockerfile:1.7

# ─── Build stage ──────────────────────────────────────────────────────────────
# Using the Alpine variant of the official Go image keeps the builder small.
# We pin to the minor version that matches go.mod; Docker resolves the latest
# patch automatically.
FROM golang:1.26-alpine AS build

WORKDIR /src

# Cache module downloads as a separate layer — they only re-run when go.sum
# changes, not on every source edit.
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source. The .dockerignore keeps this lean.
COPY . .

# CGO_ENABLED=0 produces a fully static binary so we can run on a distroless
# base that has no libc. -ldflags strips DWARF debug info and the symbol table
# to shave a few MB off the binary.
RUN CGO_ENABLED=0 GOOS=linux go build \
        -trimpath \
        -ldflags="-s -w" \
        -o /out/server \
        ./cmd/server

# ─── Runtime stage ────────────────────────────────────────────────────────────
# Distroless static: no shell, no package manager, ~2MB base. The :nonroot tag
# bakes in a non-root UID so we don't need a USER directive of our own.
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/server /server

EXPOSE 8080

ENTRYPOINT ["/server"]
