package mailer

import (
    "fmt"
    "codewax/internal/services"
)

func SendResetEmail(toEmail string, resetLink string) error {
    htmlBody := fmt.Sprintf(`
        <h1 style="text-align: center;">Password Reset</h1>
        <p style="text-align: center;">
            You requested a password reset. Click the button below to continue:
        </p>
        <div style="margin: 20px 0; text-align: center;">
            <a href="%s"
            style="background: black;
                    color: #8f8f8f;
                    padding: 12px 20px;
                    text-decoration: none;
                    display: inline-block;">
                Reset Password
            </a>
        </div>
        <p style="text-align: center;">
            If the button doesn't work, copy and paste this link:
        </p>
        <p style="word-break: break-all; text-align: center;">%s</p>
    `, resetLink, resetLink)

    return services.SendMail(toEmail, "Reset your password", htmlBody)
}