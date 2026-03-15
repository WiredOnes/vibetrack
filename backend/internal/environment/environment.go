package environment

type Environment struct {
	HostName       string `env:"HOSTNAME" validate:"required"`
	TLSCertificate string `env:"TLS_CERTIFICATE"`
	TLSKey         string `env:"TLS_KEY"`
	DBPassword     string `env:"DB_PASSWORD" validate:"required"`
}
