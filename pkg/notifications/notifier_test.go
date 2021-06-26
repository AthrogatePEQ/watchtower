package notifications_test

import (
	"fmt"
	"net/url"
	"os"
	"testing"

	"github.com/containrrr/watchtower/cmd"
	"github.com/containrrr/watchtower/internal/flags"
	"github.com/containrrr/watchtower/pkg/notifications"
	"github.com/containrrr/watchtower/pkg/types"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestActions(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Notifier Suite")
}

var _ = Describe("notifications", func() {
	Describe("the notifier", func() {
		When("only empty notifier types are provided", func() {

			command := cmd.NewRootCommand()
			flags.RegisterNotificationFlags(command)

			err := command.ParseFlags([]string{
				"--notifications",
				"shoutrrr",
			})
			Expect(err).NotTo(HaveOccurred())
			notif := notifications.NewNotifier()

			Expect(notif.String()).To(Equal("none"))
		})
	})
	Describe("the slack notifier", func() {
		builderFn := notifications.NewSlackNotifier

		When("passing a discord url to the slack notifier", func() {
			command := cmd.NewRootCommand()
			flags.RegisterNotificationFlags(command)

			channel := "123456789"
			token := "abvsihdbau"
			color := notifications.ColorInt
			title := url.QueryEscape(notifications.GetTitle())
			expected := fmt.Sprintf("discord://%s@%s?color=0x%x&colordebug=0x0&colorerror=0x0&colorinfo=0x0&colorwarn=0x0&splitlines=Yes&title=%s&username=watchtower", token, channel, color, title)
			buildArgs := func(url string) []string {
				return []string{
					"--notifications",
					"slack",
					"--notification-slack-hook-url",
					url,
				}
			}

			It("should return a discord url when using a hook url with the domain discord.com", func() {
				hookURL := fmt.Sprintf("https://%s/api/webhooks/%s/%s/slack", "discord.com", channel, token)
				testURL(builderFn, buildArgs(hookURL), expected)
			})
			It("should return a discord url when using a hook url with the domain discordapp.com", func() {
				hookURL := fmt.Sprintf("https://%s/api/webhooks/%s/%s/slack", "discordapp.com", channel, token)
				testURL(builderFn, buildArgs(hookURL), expected)
			})
		})
		When("converting a slack service config into a shoutrrr url", func() {

			It("should return the expected URL", func() {
				command := cmd.NewRootCommand()
				flags.RegisterNotificationFlags(command)

				username := "containrrrbot"
				tokenA := "aaa"
				tokenB := "bbb"
				tokenC := "ccc"
				color := url.QueryEscape(notifications.ColorHex)
				title := url.QueryEscape(notifications.GetTitle())

				hookURL := fmt.Sprintf("https://hooks.slack.com/services/%s/%s/%s", tokenA, tokenB, tokenC)
				expectedOutput := fmt.Sprintf("slack://%s@%s/%s/%s?color=%s&title=%s", username, tokenA, tokenB, tokenC, color, title)

				args := []string{
					"--notification-slack-hook-url",
					hookURL,
					"--notification-slack-identifier",
					username,
				}

				testURL(builderFn, args, expectedOutput)
			})
		})
	})

	Describe("the gotify notifier", func() {
		When("converting a gotify service config into a shoutrrr url", func() {
			builderFn := notifications.NewGotifyNotifier

			It("should return the expected URL", func() {
				command := cmd.NewRootCommand()
				flags.RegisterNotificationFlags(command)

				token := "aaa"
				host := "shoutrrr.local"
				title := url.QueryEscape(notifications.GetTitle())

				expectedOutput := fmt.Sprintf("gotify://%s/%s?title=%s", host, token, title)

				args := []string{
					"--notification-gotify-url",
					fmt.Sprintf("https://%s", host),
					"--notification-gotify-token",
					token,
				}

				testURL(builderFn, args, expectedOutput)
			})
		})
	})

	Describe("the teams notifier", func() {
		When("converting a teams service config into a shoutrrr url", func() {
			builderFn := notifications.NewMsTeamsNotifier

			It("should return the expected URL", func() {
				command := cmd.NewRootCommand()
				flags.RegisterNotificationFlags(command)

				tokenA := "11111111-4444-4444-8444-cccccccccccc@22222222-4444-4444-8444-cccccccccccc"
				tokenB := "33333333012222222222333333333344"
				tokenC := "44444444-4444-4444-8444-cccccccccccc"
				color := url.QueryEscape(notifications.ColorHex)
				title := url.QueryEscape(notifications.GetTitle())

				hookURL := fmt.Sprintf("https://outlook.office.com/webhook/%s/IncomingWebhook/%s/%s", tokenA, tokenB, tokenC)
				expectedOutput := fmt.Sprintf("teams://%s/%s/%s?color=%s&title=%s", tokenA, tokenB, tokenC, color, title)

				args := []string{
					"--notification-msteams-hook",
					hookURL,
				}

				testURL(builderFn, args, expectedOutput)
			})
		})
	})

	Describe("the email notifier", func() {

		builderFn := notifications.NewEmailNotifier

		When("converting an email service config into a shoutrrr url", func() {
			It("should set the from address in the URL", func() {
				fromAddress := "lala@example.com"
				expectedOutput := buildExpectedURL("containrrrbot", "secret-password", "mail.containrrr.dev", 25, fromAddress, "mail@example.com", "Plain")
				args := []string{
					"--notification-email-from",
					fromAddress,
					"--notification-email-to",
					"mail@example.com",
					"--notification-email-server-user",
					"containrrrbot",
					"--notification-email-server-password",
					"secret-password",
					"--notification-email-server",
					"mail.containrrr.dev",
				}
				testURL(builderFn, args, expectedOutput)
			})

			It("should return the expected URL", func() {

				fromAddress := "sender@example.com"
				toAddress := "receiver@example.com"
				expectedOutput := buildExpectedURL("containrrrbot", "secret-password", "mail.containrrr.dev", 25, fromAddress, toAddress, "Plain")

				args := []string{
					"--notification-email-from",
					fromAddress,
					"--notification-email-to",
					toAddress,
					"--notification-email-server-user",
					"containrrrbot",
					"--notification-email-server-password",
					"secret-password",
					"--notification-email-server",
					"mail.containrrr.dev",
				}

				testURL(builderFn, args, expectedOutput)
			})
		})
	})
})

func buildExpectedURL(username string, password string, host string, port int, from string, to string, auth string) string {
	hostname, err := os.Hostname()
	Expect(err).NotTo(HaveOccurred())

	subject := fmt.Sprintf("Watchtower updates on %s", hostname)

	var template = "smtp://%s:%s@%s:%d/?auth=%s&fromaddress=%s&fromname=Watchtower&subject=%s&toaddresses=%s"
	return fmt.Sprintf(template,
		url.QueryEscape(username),
		url.QueryEscape(password),
		host, port, auth,
		url.QueryEscape(from),
		url.QueryEscape(subject),
		url.QueryEscape(to))
}

type builderFn = func() types.ConvertibleNotifier

func testURL(builder builderFn, args []string, expectedURL string) {

	command := cmd.NewRootCommand()
	flags.RegisterNotificationFlags(command)

	err := command.ParseFlags(args)
	Expect(err).NotTo(HaveOccurred())

	notifier := builder()
	actualURL, err := notifier.GetURL()

	Expect(err).NotTo(HaveOccurred())

	Expect(actualURL).To(Equal(expectedURL))
}
