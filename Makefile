PKG_PATH := ./charts/repo

chart:
	cr package ./charts/k8s-pvc-usage --package-path $(PKG_PATH)
	cr upload --package-path $(PKG_PATH) \
		--owner codekoala \
		--git-repo k8s-pvc-usage \
		--packages-with-index \
		--skip-existing \
		--push
	cr index --package-path $(PKG_PATH) --index-path $(PKG_PATH) \
		--owner codekoala \
		--git-repo k8s-pvc-usage \
		--packages-with-index \
		--push
