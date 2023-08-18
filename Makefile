chart:
	cr package ./charts/k8s-pvc-usage --package-path ./charts/repo
	cr upload \
		--owner codekoala \
		--git-repo k8s-pvc-usage \
		--package-path ./charts/repo \
		--packages-with-index \
		--skip-existing \
		--push
	cr index \
		--owner codekoala \
		--git-repo k8s-pvc-usage \
		--packages-with-index \
		--index-path ./charts/repo \
		--push
