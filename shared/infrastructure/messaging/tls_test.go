package messaging

import (
	"net/url"
	"testing"

	"github.com/hycat-team/littlecafe-backend/shared/config"
)

func TestAMQPURLEscapesCredentials(t *testing.T) {
	cfg := &config.Config{}
	cfg.RabbitMQ.Host = "rabbitmq"
	cfg.RabbitMQ.Port = 5672
	cfg.RabbitMQ.User = "littlecafe"
	cfg.RabbitMQ.Password = "p/a+s=s@:x y"
	cfg.RabbitMQ.Vhost = "/"

	raw := amqpURL(cfg)
	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("amqpURL() produced unparsable URL %q: %v", raw, err)
	}
	if parsed.Scheme != "amqp" {
		t.Fatalf("scheme = %q, want amqp", parsed.Scheme)
	}
	if parsed.Host != "rabbitmq:5672" {
		t.Fatalf("host = %q, want rabbitmq:5672", parsed.Host)
	}
	if parsed.Path != "/" {
		t.Fatalf("path = %q, want /", parsed.Path)
	}
	if parsed.User.Username() != cfg.RabbitMQ.User {
		t.Fatalf("username = %q, want %q", parsed.User.Username(), cfg.RabbitMQ.User)
	}
	password, ok := parsed.User.Password()
	if !ok || password != cfg.RabbitMQ.Password {
		t.Fatalf("password = %q (ok=%v), want %q", password, ok, cfg.RabbitMQ.Password)
	}
}

func TestAMQPURLUsesTLSAndCustomVhost(t *testing.T) {
	cfg := &config.Config{}
	cfg.RabbitMQ.Host = "rabbitmq"
	cfg.RabbitMQ.Port = 5671
	cfg.RabbitMQ.User = "littlecafe"
	cfg.RabbitMQ.Password = "secret"
	cfg.RabbitMQ.Vhost = "littlecafe"
	cfg.RabbitMQ.TLSEnabled = true

	raw := amqpURL(cfg)
	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("amqpURL() produced unparsable URL %q: %v", raw, err)
	}
	if parsed.Scheme != "amqps" {
		t.Fatalf("scheme = %q, want amqps", parsed.Scheme)
	}
	if parsed.Host != "rabbitmq:5671" {
		t.Fatalf("host = %q, want rabbitmq:5671", parsed.Host)
	}
	if parsed.Path != "/littlecafe" {
		t.Fatalf("path = %q, want /littlecafe", parsed.Path)
	}
}
