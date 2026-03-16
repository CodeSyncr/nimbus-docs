# Mail

> **Email sending made simple** — SMTP and native API drivers for major email providers with a clean, unified API.

---

## Introduction

Nimbus provides a unified mail API for sending emails through SMTP or native HTTP APIs. Whether you're using a local SMTP server for development, Amazon SES for transactional emails, or SendGrid for marketing campaigns, the API remains the same.

Features:

- **Driver-based architecture** — Common `Driver` interface for all providers
- **SMTP drivers** — SMTP, Amazon SES (SMTP), Mailgun (SMTP), SendGrid (SMTP), Postmark
- **Native API drivers** — SendGrid v3 API, Mailgun API, SES API, Resend API (no SMTP required)
- **Simple message API** — `To`, `From`, `Subject`, `Body`, `HTML`, `CC`, `BCC`, `Attachments`
- **Global default driver** — Set once at boot, use everywhere
- **Queue integration** — Dispatch emails as background jobs

---

## Configuration

```env
MAIL_DRIVER=smtp             # smtp | ses | mailgun | sendgrid | postmark
MAIL_HOST=localhost
MAIL_PORT=1025
MAIL_USERNAME=
MAIL_PASSWORD=
MAIL_FROM=noreply@example.com
MAIL_ENCRYPTION=tls
```

### Boot Configuration

```go
// bin/server.go
func bootMail() {
    mail.Default = mail.NewSMTPDriver(
        os.Getenv("MAIL_HOST")+":"+os.Getenv("MAIL_PORT"),
        smtp.PlainAuth("", os.Getenv("MAIL_USERNAME"), os.Getenv("MAIL_PASSWORD"), os.Getenv("MAIL_HOST")),
        os.Getenv("MAIL_FROM"),
    )
}
```

---

## Driver Interface

```go
type Driver interface {
    Send(m *Message) error
}
```

All mail drivers implement a single `Send` method. Nimbus wraps provider-specific SMTP configurations into pre-built constructors.

---

## Message Structure

```go
type Message struct {
    From    string   // Sender email address
    To      []string // Recipient email addresses
    Subject string   // Email subject line
    Body    string   // Email body (plain text or HTML)
    HTML    bool     // If true, body is sent as HTML
}
```

---

## Sending Mail

### Basic Text Email

```go
import "github.com/CodeSyncr/nimbus/mail"

err := mail.Send(&mail.Message{
    To:      []string{"user@example.com"},
    Subject: "Welcome to our platform",
    Body:    "Thank you for signing up!",
})
```

### HTML Email

```go
err := mail.Send(&mail.Message{
    To:      []string{"user@example.com"},
    Subject: "Your Order Confirmation",
    Body: `
        <h1>Order Confirmed!</h1>
        <p>Thank you for your purchase.</p>
        <p>Order #12345 has been confirmed and will ship within 2 business days.</p>
    `,
    HTML: true,
})
```

### Multiple Recipients

```go
err := mail.Send(&mail.Message{
    To:      []string{"alice@example.com", "bob@example.com", "carol@example.com"},
    Subject: "Team Update",
    Body:    "Weekly status report attached.",
})
```

### Custom From Address

```go
err := mail.Send(&mail.Message{
    From:    "support@myapp.com",
    To:      []string{"customer@example.com"},
    Subject: "Support Ticket #4567",
    Body:    "Your issue has been resolved.",
})
```

---

## Drivers

### SMTP (Generic)

Works with any SMTP server:

```go
driver := mail.NewSMTPDriver(
    "smtp.example.com:587",
    smtp.PlainAuth("", "user", "pass", "smtp.example.com"),
    "noreply@example.com",
)
```

### Amazon SES

```go
driver := mail.NewSESDriver(
    "email-smtp.us-east-1.amazonaws.com:587",
    smtp.PlainAuth("", os.Getenv("AWS_SES_USER"), os.Getenv("AWS_SES_PASS"), "email-smtp.us-east-1.amazonaws.com"),
    "noreply@yourdomain.com",
)
```

### Mailgun (SMTP)

```go
driver := mail.NewMailgunSMTPDriver(
    "smtp.mailgun.org:587",
    smtp.PlainAuth("", "postmaster@mg.yourdomain.com", os.Getenv("MAILGUN_PASS"), "smtp.mailgun.org"),
    "noreply@yourdomain.com",
)
```

### SendGrid (SMTP)

```go
driver := mail.NewSendGridSMTPDriver(
    "smtp.sendgrid.net:587",
    smtp.PlainAuth("", "apikey", os.Getenv("SENDGRID_API_KEY"), "smtp.sendgrid.net"),
    "noreply@yourdomain.com",
)
```

### Postmark

```go
driver := mail.NewPostmarkDriver(
    "smtp.postmarkapp.com:587",
    smtp.PlainAuth("", os.Getenv("POSTMARK_TOKEN"), os.Getenv("POSTMARK_TOKEN"), "smtp.postmarkapp.com"),
    "noreply@yourdomain.com",
)
```

---

## Native API Drivers

Native API drivers communicate directly with the provider's HTTP API — no SMTP server needed. These are recommended for production deployments.

### SendGrid (API)

```go
driver := mail.NewSendGridDriver(
    os.Getenv("SENDGRID_API_KEY"),
    "noreply@yourdomain.com",
)
mail.Default = driver
```

### Mailgun (API)

```go
driver := mail.NewMailgunAPIDriver(
    "your-domain.com",
    os.Getenv("MAILGUN_API_KEY"),
    "noreply@your-domain.com",
)
mail.Default = driver
```

### SES (API)

```go
driver := mail.NewSESAPIDriver(
    "us-east-1",
    os.Getenv("AWS_ACCESS_KEY_ID"),
    os.Getenv("AWS_SECRET_ACCESS_KEY"),
    "noreply@yourdomain.com",
)
mail.Default = driver
```

### Resend (API)

```go
driver := mail.NewResendDriver(
    os.Getenv("RESEND_API_KEY"),
    "noreply@yourdomain.com",
)
mail.Default = driver
```

---

## Sending Mail via Queue

For non-blocking email delivery, dispatch mail as a background job:

```go
// app/jobs/send_welcome_email.go
type SendWelcomeEmail struct {
    UserID uint
    Email  string
}

func (j *SendWelcomeEmail) Handle(ctx context.Context) error {
    return mail.Send(&mail.Message{
        To:      []string{j.Email},
        Subject: "Welcome to MyApp!",
        Body: fmt.Sprintf(`
            <h1>Welcome!</h1>
            <p>Hi there! Your account has been created successfully.</p>
            <p>Get started by visiting your dashboard.</p>
        `),
        HTML: true,
    })
}

func (j *SendWelcomeEmail) Failed(ctx context.Context, err error) {
    logger.Error("welcome email failed", "email", j.Email, "error", err)
}
```

Dispatch from a controller:

```go
func (ctrl *AuthController) Register(c *http.Context) error {
    // ... create user ...

    queue.Dispatch(&jobs.SendWelcomeEmail{
        UserID: user.ID,
        Email:  user.Email,
    }).Dispatch(c.Request.Context())

    return c.JSON(201, user)
}
```

---

## Real-Life Examples

### Notification System

```go
// app/jobs/send_notification_email.go
type SendNotificationEmail struct {
    Email   string
    Title   string
    Message string
    Action  string // URL
}

func (j *SendNotificationEmail) Handle(ctx context.Context) error {
    body := fmt.Sprintf(`
        <div style="font-family: sans-serif; max-width: 600px; margin: 0 auto;">
            <h2>%s</h2>
            <p>%s</p>
            <a href="%s" style="display: inline-block; padding: 12px 24px; 
               background: #3b82f6; color: white; text-decoration: none; 
               border-radius: 6px;">View Details</a>
        </div>
    `, j.Title, j.Message, j.Action)

    return mail.Send(&mail.Message{
        To:      []string{j.Email},
        Subject: j.Title,
        Body:    body,
        HTML:    true,
    })
}
```

### Password Reset

```go
func sendPasswordResetEmail(email, token string) error {
    resetURL := fmt.Sprintf("https://myapp.com/reset-password?token=%s", token)

    return mail.Send(&mail.Message{
        To:      []string{email},
        Subject: "Reset Your Password",
        Body: fmt.Sprintf(`
            <h2>Password Reset</h2>
            <p>Click the link below to reset your password. This link expires in 1 hour.</p>
            <a href="%s">Reset Password</a>
            <p style="color: #666; font-size: 12px;">If you didn't request this, ignore this email.</p>
        `, resetURL),
        HTML: true,
    })
}
```

### Order Confirmation

```go
func sendOrderConfirmation(order Order) error {
    var itemsHTML string
    for _, item := range order.Items {
        itemsHTML += fmt.Sprintf(
            `<tr><td>%s</td><td>%d</td><td>$%.2f</td></tr>`,
            item.Name, item.Quantity, item.Price,
        )
    }

    return mail.Send(&mail.Message{
        From:    "orders@myshop.com",
        To:      []string{order.User.Email},
        Subject: fmt.Sprintf("Order #%d Confirmed", order.ID),
        Body: fmt.Sprintf(`
            <h2>Order Confirmed!</h2>
            <p>Order #%d has been placed successfully.</p>
            <table>
                <tr><th>Item</th><th>Qty</th><th>Price</th></tr>
                %s
            </table>
            <p><strong>Total: $%.2f</strong></p>
        `, order.ID, itemsHTML, order.Total),
        HTML: true,
    })
}
```

---

## Development Setup

For local development, use [MailHog](https://github.com/mailhog/MailHog) or [Mailpit](https://github.com/axllent/mailpit):

```env
MAIL_HOST=localhost
MAIL_PORT=1025
MAIL_USERNAME=
MAIL_PASSWORD=
MAIL_FROM=dev@localhost
```

```bash
# Start Mailpit
brew install mailpit
mailpit
# Web UI at http://localhost:8025
# SMTP at localhost:1025
```

---

## Best Practices

1. **Always queue emails** — Never send during HTTP request; use background jobs
2. **Use HTML templates** — Build consistent, branded emails
3. **Include plain-text fallback** — Some clients don't render HTML
4. **Handle failures** — Implement `Failed()` on email jobs for retry logging
5. **Validate email addresses** — Before queueing, check format
6. **Rate limit sends** — Respect provider limits (SES: 14/sec, SendGrid: 100/sec)
7. **Use environment variables** — Never hardcode API keys or credentials

**Next:** [Scheduler](15-scheduler.md) →
