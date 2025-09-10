# Atlas configuration for schema management
env "local" {
  src = "file://internal/platform/database/migrations"
  url = "postgres://testuser:testpass@localhost:5432/image_gallery_test?sslmode=disable"
  dev = "docker://postgres/15/test?search_path=public"
}

env "test" {
  src = "file://internal/platform/database/migrations"
  url = "postgres://testuser:testpass@localhost:5432/image_gallery_test?sslmode=disable"
  dev = "docker://postgres/15/test?search_path=public"
  
  # Enable migration testing
  test {
    schema {
      src = "file://internal/platform/database/migrations"
    }
  }
}

env "prod" {
  src = "file://internal/platform/database/migrations"
  url = env("DATABASE_URL")
  
  # Production migration settings
  migration {
    dir = "file://internal/platform/database/migrations"
    format = "sql"
  }
  
  # Enable migration linting
  lint {
    log = <<EOS
      {{- range $f := .Files }}
        {{- $f.Name }}:
        {{- range $r := $f.Reports }}
          - {{ $r.Text }}
        {{- end }}
      {{- end }}
    EOS
  }
}