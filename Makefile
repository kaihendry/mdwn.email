deploy:
	apex deploy -r us-west-2

logs:
	apex logs hello -r us-west-2

upload:
	aws --region us-west-2 --profile mine \
		s3 sync --storage-class REDUCED_REDUNDANCY --acl public-read index/ s3://mdwn-web/
	@echo http://mdwn-web.s3-website-us-west-2.amazonaws.com
	@aws --profile mine cloudfront create-invalidation --distribution-id EZWNPV8ZIIIJI --invalidation-batch "{ \"Paths\": { \"Quantity\": 1, \"Items\": [ \"/*\" ] }, \"CallerReference\": \"$(shell date +%s)\" }"
