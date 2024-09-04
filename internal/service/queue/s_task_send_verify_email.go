package queue

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go-todolist-grpc/internal/config"
	"go-todolist-grpc/internal/model"
	"go-todolist-grpc/internal/pkg/db"
	"go-todolist-grpc/internal/pkg/log"
	"go-todolist-grpc/internal/pkg/mail"
	"go-todolist-grpc/internal/pkg/util"
	"html/template"
	"time"

	"github.com/hibiken/asynq"
)

const TaskSendVerifyEmail = "send_verify_email"
const htmlTemplate = `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Email Verification</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            line-height: 1.6;
            color: #333;
        }
        .container {
            max-width: 600px;
            margin: 0 auto;
            padding: 20px;
        }
        .button {
            display: inline-block;
            padding: 10px 20px;
            background-color: #007bff;
            color: #ffffff;
            text-decoration: none;
            border-radius: 5px;
        }
    </style>
</head>
<body>
    <div class="container">
        <h2>Hello {{.Username}},</h2>
        <p>Thank you for registering with us!</p>
        <p>Please click the button below to verify your email address:</p>
        <p>
            <a href="{{.VerifyURL}}" class="button">Verify Email</a>
        </p>
        <p>If the button doesn't work, you can copy and paste this link into your browser:</p>
        <p>{{.VerifyURL}}</p>
        <p>Best regards,<br>Your Team</p>
    </div>
</body>
</html>
`

type PayloadSendVerifyEmail struct {
	UserId int `json:"user_id"`
}

func (rtd *RedisTaskDistributor) DistributeTaskSendVerifyEmail(ctx context.Context, payload *PayloadSendVerifyEmail, opts ...asynq.Option) error {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal task payload: %w", err)
	}

	task := asynq.NewTask(TaskSendVerifyEmail, jsonPayload, opts...)
	info, err := rtd.client.EnqueueContext(ctx, task)
	if err != nil {
		return err
	}

	log.Info.Printf("enqueued task - type: %s, payload (userID): %s, queue: %s, max_retry: %d", task.Type(), string(task.Payload()), info.Queue, info.MaxRetry)
	return nil
}

func (p *RedisTaskProcessor) ProcessTaskSendVerifyEmail(ctx context.Context, task *asynq.Task) error {
	type mailContent struct {
		Username  string
		VerifyURL string
	}

	cnf := config.Get()
	conn := db.GetConn()
	payload := PayloadSendVerifyEmail{}
	if err := json.Unmarshal(task.Payload(), &payload); err != nil {
		return fmt.Errorf("failed to unmarshal payload: %w", asynq.SkipRetry)
	}

	// Get the user info
	getUser := model.GetUserByID(conn, payload.UserId)
	if getUser == nil {
		return errors.New("[send verify email] - user ID not found")
	}

	now := time.Now().UTC()
	createVerifyEmail, createVerifyEmailErr := model.CreateVerifyEmail(conn, &model.VerifyEmailFieldValues{
		UserId:     model.GiveColInt(getUser.ID),
		Username:   model.GiveColString(getUser.Username),
		Email:      model.GiveColString(getUser.Email),
		SecretCode: model.GiveColString(util.RandomString(32)),
		IsUsed:     model.GiveColBool(false),
		ExpiredAt:  model.GiveColTime(now.Add(1 * time.Hour)),
		CreatedAt:  model.GiveColTime(now),
		UpdatedAt:  model.GiveColTime(now),
	})
	if createVerifyEmailErr != nil {
		return fmt.Errorf("failed to create verify email: %w", createVerifyEmailErr)
	}

	// Define email content
	verifyUrl := fmt.Sprintf(
		"http://localhost:%s/v1/user/verify_email?id=%d&secret_code=%s",
		cnf.HttpServerPort,
		createVerifyEmail.ID.Val,
		createVerifyEmail.SecretCode.Val,
	)
	data := mailContent{
		Username:  createVerifyEmail.Username.Val,
		VerifyURL: verifyUrl,
	}

	tmpl, tmplErr := template.New("emailTemplate").Parse(htmlTemplate)
	if tmplErr != nil {
		return fmt.Errorf("error parsing template: %v", tmplErr)
	}

	htmlStr := bytes.Buffer{}
	if err := tmpl.Execute(&htmlStr, data); err != nil {
		return fmt.Errorf("error executing template: %v", err)
	}

	if err := mail.SendEmail(
		cnf,
		fmt.Sprintf("%s <%s>", cnf.EmailSenderName, cnf.EmailSenderAddress), // sender
		[]string{createVerifyEmail.Email.Val},                               // recipient
		[]string{},                                                          // bccs
		"Welcome to Go-Todolist-gRPC",                                       // subject
		htmlStr.String(),                                                    // htmlBody
		"",                                                                  // textBody
	); err != nil {
		log.Error.Printf("sent verify email error: %v", err)
		return err
	}
	log.Info.Printf("processed task - type: %s, payload (userID): %s, email: %s", task.Type(), string(task.Payload()), getUser.Email)

	return nil
}
