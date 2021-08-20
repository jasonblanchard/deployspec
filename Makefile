dockerbuild:
	docker build . -t jasonblanchard/deployspec

dockerpush: dockerbuild
	docker push jasonblanchard/deployspec

ecrlogin:
	aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin XXX.dkr.ecr.us-east-1.amazonaws.com