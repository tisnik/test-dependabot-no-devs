#include <getopt.h>
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>

/**
 * Parse command-line options (short and long forms), print each recognized option
 * with its argument (if any), list any remaining non-option arguments, and exit.
 *
 * @param argc Number of command-line arguments.
 * @param argv Array of command-line argument strings.
 * @returns EXIT_SUCCESS on successful completion.
 */
int main(int argc, char *argv[]) {
  int c;

  while (1) {
    int option_index = 0;
    static struct option long_options[] = {
        {"fractal", required_argument, NULL, 0},
        {"pattern", no_argument, NULL, 0},
        {"filter", required_argument, NULL, 0},
        {"palette", required_argument, NULL, 0},
        {"width", required_argument, NULL, 0},
        {"height", required_argument, NULL, 0},
        {"maxiter", required_argument, NULL, 0},
        {"output", required_argument, NULL, 0},
        {"first", required_argument, NULL, 0},
        {"second", required_argument, NULL, 0},
        {"third", required_argument, NULL, 0},
        {"verbose", no_argument, NULL, 0},
        {"help", no_argument, NULL, 0},
        {NULL, 0, NULL, 0}};

 /* "c" and "d" arguments required parameter */
    c  =  getopt_long(argc,  argv,   "abc:d:",   long_options,   &option_index);

    if (c == -1)
      break;

    switch (c) {
    case 0:
      printf("option %s", long_options[option_index].name);
      if (optarg)
        printf(" with arg %s", optarg);
      printf("\n");
      break;

    case 'a':
      printf("option a\n");
      break;

    case 'b':
      printf("option b\n");
      break;

    case 'c':
      printf("option c with value '%s'\n", optarg);
      break;

    case 'd':
      printf("option d with value '%s'\n", optarg);
      break;

    case '?':
      printf("PROBLEM");
      break;

    default:
      printf("?? getopt returned character code 0%o ??\n", c);
    }
  }

  if (optind < argc) {
    printf("non-option ARGV-elements: ");
    while (optind < argc)
      printf("%s ", argv[optind++]);
    printf("\n");
  }

  exit(EXIT_SUCCESS);
}