#include <dlfcn.h>
#include <stdio.h>
#include <stdlib.h>

int main() {
	void *handle = dlopen("./test-dynamic.so", RTLD_LAZY);
	if (!handle) {
		fprintf(stderr, "dlopen: %s\n", dlerror());
		return 1;
	}

	int (*go_test_main_with_args)(int, const char**) = dlsym(handle, "go_test_main_with_args");
	char *err = dlerror();
	if (err) {
		fprintf(stderr, "dlsym: %s\n", err);
		dlclose(handle);
		return 1;
	}

	int ret = go_test_main_with_args(2, (const char*[]){"<none>", "-test.v"});
	fprintf(stderr, "Go test main returned: %d\n", ret);
	return ret;

	dlclose(handle);
	return 0;
}
