package config

type Configuration struct {
	Server  *Server  `environment:"SERVER"`
	Storage *Storage `environment:"STORAGE"`
}

type Server struct {
	Addr                    string   `environment:"ADDR"`
	ReadHeaderTimeoutMillis int      `environment:"READ_HEADER_TIMEOUT_MILLIS"`
	CorsAllowedOrigins      []string `environment:"CORS_ALLOWED_ORIGINS"`
}

type Storage struct {
	Type     StorageType `environment:"TYPE"`
	Postgres *Postgres   `environment:"POSTGRES"`
}

type StorageType string

const (
	StorageTypeMemory   StorageType = "memory"
	StorageTypePostgres StorageType = "postgres"
)

// Postgres configures the Postgres connection.
// These options come from https://godoc.org/github.com/lib/pq#hdr-Connection_String_Parameters
type Postgres struct {
	Database string `environment:"DBNAME"`
	Host     string `environment:"HOST"`
	Port     int    `environment:"PORT"`
	User     string `environment:"USER"`
	Password string `environment:"PASSWORD"`
}

func (cfg *Configuration) logValues() {
	cfg.Server.logValues()
	cfg.Storage.logValues()
}

func (cfg *Configuration) validate() []error {
	errs := cfg.Server.validate()
	errs = append(errs, cfg.Storage.validate()...)
	return errs
}
