CREATE TABLE employees (
    id          BIGSERIAL PRIMARY KEY,
    name        TEXT NOT NULL,
    hourly_rate BIGINT NOT NULL CHECK (hourly_rate >= 0),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE payslips (
    id              BIGSERIAL PRIMARY KEY,
    employee_id     BIGINT NOT NULL REFERENCES employees(id) ON DELETE RESTRICT,
    hours_worked    DOUBLE PRECISION NOT NULL CHECK (hours_worked >= 0),
    gross_pay       BIGINT NOT NULL,
    federal_tax     BIGINT NOT NULL,
    social_security BIGINT NOT NULL,
    medicare        BIGINT NOT NULL,
    net_pay         BIGINT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX payslips_employee_id_idx ON payslips(employee_id);
