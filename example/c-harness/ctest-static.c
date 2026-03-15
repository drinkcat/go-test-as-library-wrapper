#include <stdio.h>

int go_test_main_with_args(int argc, const char **argv);

int main() {
	int ret = go_test_main_with_args(2, (const char*[]){"<none>", "-test.v"});
	fprintf(stderr,"Go test main returned: %d\n", ret);
	return ret;
}
