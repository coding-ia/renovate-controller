package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"os"
	"testing"
)

var body = `{"action":"removed","installation":{"id":53773974,"account":{"login":"coding-ia","id":137450628,"node_id":"U_kgDOCDFUhA","avatar_url":"https://avatars.githubusercontent.com/u/137450628?v=4","gravatar_id":"","url":"https://api.github.com/users/coding-ia","html_url":"https://github.com/coding-ia","followers_url":"https://api.github.com/users/coding-ia/followers","following_url":"https://api.github.com/users/coding-ia/following{/other_user}","gists_url":"https://api.github.com/users/coding-ia/gists{/gist_id}","starred_url":"https://api.github.com/users/coding-ia/starred{/owner}{/repo}","subscriptions_url":"https://api.github.com/users/coding-ia/subscriptions","organizations_url":"https://api.github.com/users/coding-ia/orgs","repos_url":"https://api.github.com/users/coding-ia/repos","events_url":"https://api.github.com/users/coding-ia/events{/privacy}","received_events_url":"https://api.github.com/users/coding-ia/received_events","type":"User","site_admin":false},"repository_selection":"selected","access_tokens_url":"https://api.github.com/app/installations/53773974/access_tokens","repositories_url":"https://api.github.com/installation/repositories","html_url":"https://github.com/settings/installations/53773974","app_id":969806,"app_slug":"coding-ia-test","target_id":137450628,"target_type":"User","permissions":{"administration":"read","checks":"write","contents":"write","issues":"write","metadata":"read","pull_requests":"write","statuses":"write","vulnerability_alerts":"read","workflows":"write"},"events":[],"created_at":"2024-08-13T20:10:27.000-07:00","updated_at":"2024-08-18T16:54:47.000-07:00","single_file_name":null,"has_multiple_single_files":false,"single_file_paths":[],"suspended_by":null,"suspended_at":null},"repository_selection":"selected","repositories_added":[],"repositories_removed":[{"id":842589580,"node_id":"R_kgDOMjjljA","name":"renovate-controller","full_name":"coding-ia/renovate-controller","private":false}],"requester":null,"sender":{"login":"coding-ia","id":137450628,"node_id":"U_kgDOCDFUhA","avatar_url":"https://avatars.githubusercontent.com/u/137450628?v=4","gravatar_id":"","url":"https://api.github.com/users/coding-ia","html_url":"https://github.com/coding-ia","followers_url":"https://api.github.com/users/coding-ia/followers","following_url":"https://api.github.com/users/coding-ia/following{/other_user}","gists_url":"https://api.github.com/users/coding-ia/gists{/gist_id}","starred_url":"https://api.github.com/users/coding-ia/starred{/owner}{/repo}","subscriptions_url":"https://api.github.com/users/coding-ia/subscriptions","organizations_url":"https://api.github.com/users/coding-ia/orgs","repos_url":"https://api.github.com/users/coding-ia/repos","events_url":"https://api.github.com/users/coding-ia/events{/privacy}","received_events_url":"https://api.github.com/users/coding-ia/received_events","type":"User","site_admin":false}}`

func TestAPIGatewaySignature_Valid(t *testing.T) {
	ctx := context.Background()
	_ = os.Setenv("AWS_ECS_TASK_PUBLIC_IP", "True")
	err := os.Setenv("WEBHOOK_SECRET", "abc123")
	if err != nil {
		t.Fail()
	}

	event := events.APIGatewayProxyRequest{
		Body: body,
		Headers: map[string]string{
			"X-GitHub-Event":      "installation_repositories",
			"X-Hub-Signature-256": "sha256=0b23aef22b3f4e701fcd439f574384cbbcf17b300f50ecc99a5469b0101fbbc8",
		},
	}

	response, err := handler(ctx, event)
	if err != nil {
		t.Fail()
	}
	if response.Body != "" {
		t.Fail()
	}
}

func TestAPIGatewaySignature_InValid(t *testing.T) {
	ctx := context.Background()
	_ = os.Setenv("AWS_ECS_TASK_PUBLIC_IP", "True")
	err := os.Setenv("WEBHOOK_SECRET", "abc123")
	if err != nil {
		t.Fail()
	}

	event := events.APIGatewayProxyRequest{
		Body: body,
		Headers: map[string]string{
			"X-GitHub-Event":      "installation_repositories",
			"X-Hub-Signature-256": "sha256=26754c8b0afb6ac70d13f5bcb1b90b21ff2c8266d3da15d18d37f9d0f300a6a1",
		},
	}

	response, err := handler(ctx, event)
	if err != nil {
		t.Fail()
	}
	if response.Body != "{\"error\": \"Invalid signature\"}" {
		t.Fail()
	}
}
