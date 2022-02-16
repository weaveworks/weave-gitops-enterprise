# Zone delegation has been setup manually 
# in the Weave Cloud dev account
locals {
  zone_id =  "Z038735537FBV7QQ5O394"
  zone_name ="wge.dev.weave.works"
}

resource "aws_route53_record" "demo_02_ingress" {
  zone_id = local.zone_id
  name    = "demo-02"
  type    = "A"
  ttl     = "300"
  records = ["35.228.235.99"]
}

resource "aws_route53_record" "dex_01_ingress" {
  zone_id = local.zone_id
  name    = "dex-01"
  type    = "A"
  ttl     = "300"
  records = ["35.228.83.196"]
}