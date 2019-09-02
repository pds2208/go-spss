
#include <stdio.h>
#include <string.h>

#include <R.h>
#include <Rembedded.h>
#include <Rinternals.h>

void source(const char *name) {
  SEXP e;

  PROTECT(e = lang2(install("source"), mkString(name)));
  R_tryEval(e, R_GlobalEnv, NULL);
  UNPROTECT(1);
}

void initialise() {
  // Intialize the R environment.
  int r_argc = 3;
  char *r_argv[] = {"R", "--silent", "--no-save"};
  Rf_initEmbeddedR(r_argc, r_argv);
}

void load_r_source(const char *s) { source(s); }

void free_r() { Rf_endEmbeddedR(0); }
