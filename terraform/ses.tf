# This file creates the SES email sending service for completed videos

resource "aws_ses_configuration_set" "ses_configuration_videos" {
  name = "zircon_video_configuration_set"
  delivery_options {
    tls_policy = "Require"
  }
}

resource "aws_ses_domain_identity" "zircon_domain_identity" {
  domain = var.DOMAIN
}

resource "aws_ses_domain_dkim" "zircon_dkim" {
  domain = var.DOMAIN
}

resource "aws_ses_template" "zircon_job_complete_template" {
  name    = "zircon_job_complete_template"
  subject = "[Zircon] Your {{VideoTitle}} Video for {{Subject}}!"
  html    = file("${path.module}/../backend/pkg/sesClient/jobTemplate.html")
  text    = "Your requested video is ready! You can view it by opening this link in your browser:\nhttps://www.zircon.socialcoding.net/assets/{{EntryID}}/{{VideoType}}.mp4"
}
