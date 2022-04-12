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
  records = ["35.228.32.242"]
}

resource "aws_route53_record" "demo_02_ingress" {
  zone_id = local.zone_id
  name    = "demo-02"
  type    = "A"
  ttl     = "300"
  records = ["35.228.213.125"]
}

resource "aws_route53_record" "demo_03_ingress" {
  zone_id = local.zone_id
  name    = "demo-03"
  type    = "A"
  ttl     = "300"
  records = ["34.88.94.216"]
}

resource "aws_route53_record" "demo_04_ingress" {
  zone_id = local.zone_id
  name    = "demo-04"
  type    = "CNAME"
  ttl     = "60"
  records = ["a5e80ee587120483c9a82036d8893409-1250835218.eu-west-1.elb.amazonaws.com"]
}

resource "aws_route53_record" "dex_01_ingress" {
  zone_id = local.zone_id
  name    = "dex-01"
  type    = "A"
  ttl     = "300"
  records = ["35.228.83.196"]
}