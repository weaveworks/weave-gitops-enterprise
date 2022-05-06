# Zone delegation has been setup manually 
# in the Weave Cloud dev account
locals {
  zone_id =  "Z038735537FBV7QQ5O394"
  zone_name ="wge.dev.weave.works"
}

resource "aws_route53_record" "demo_01_ingress" {
  zone_id = local.zone_id
  name    = "demo-01"
  type    = "A"
  ttl     = "300"
  records = ["34.88.95.15"]
}

resource "aws_route53_record" "demo_02_ingress" {
  zone_id = local.zone_id
  name    = "demo-02"
  type    = "A"
  ttl     = "300"
  records = ["35.228.213.125"]
}

resource "aws_route53_record" "dex_01_ingress" {
  zone_id = local.zone_id
  name    = "dex-01"
  type    = "A"
  ttl     = "300"
  records = ["35.228.83.196"]
}

resource "aws_route53_record" "charts_ingress" {
  zone_id = local.zone_id
  name    = "charts"
  type    = "A"
  ttl     = "300"
  records = ["35.228.83.196"]
}