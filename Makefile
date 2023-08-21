IMG := codekoala/k8s-pvc-usage
VERSION := 0.1.2
PKG_PATH := ./charts/repo

chart:
	cr package ./charts/k8s-pvc-usage --package-path $(PKG_PATH)
	cr upload --package-path $(PKG_PATH) \
		--owner codekoala \
		--git-repo k8s-pvc-usage \
		--packages-with-index \
		--skip-existing
	cr index --package-path $(PKG_PATH) --index-path $(PKG_PATH) \
		--owner codekoala \
		--git-repo k8s-pvc-usage \
		--packages-with-index

debug:
	docker build --target debug --tag $(IMG):$(VERSION)-debug .
	docker push $(IMG):$(VERSION)-debug

image:
	docker build --tag $(IMG):$(VERSION) .
	docker push $(IMG):$(VERSION)
