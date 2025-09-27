# Atlas configuration for schema management
env "local" {
  src = "file://internal/platform/database/migrations"
  url = "postgres://testuser:testpass@localhost:5432/image_gallery_test?sslmode=disable"
  dev = "docker://postgres/15/test?search_path=public"

  migration {
    dir = "file://internal/platform/database/migrations"
  }
}

env "test" {
  src = "file://internal/platform/database/migrations"
  url = getenv("TEST_DATABASE_URL")
  dev = "docker://postgres/15/test?search_path=public"

  migration {
    dir = "file://internal/platform/database/migrations"
  }
}

env "prod" {
  src = "file://internal/platform/database/migrations"
  url = getenv("DATABASE_URL")

  # Production migration settings
  migration {
    dir = "file://internal/platform/database/migrations"
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

env "k8s" {
  src = "file://migrations"  # Path inside ConfigMap
  url = getenv("DATABASE_URL")

  # Kubernetes migration settings
  migration {
    dir = "file://migrations"
  }

  # Production safety policies
  lint {
    destructive {
      error = true  # Block destructive changes in Kubernetes
    }
    log = <<EOS
      {{- range $f := .Files }}
        {{- $f.Name }}:
        {{- range $r := $f.Reports }}
          - {{ $r.Text }}
        {{- end }}
      {{- end }}
    EOS
  }

  # Enable diff policies for safer migrations
  diff {
    # Skip destructive changes by default
    skip {
      drop_schema = true
      drop_table = true
    }
  }
}