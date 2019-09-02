
#include <stdio.h>
#include <string.h>

#include <R.h>
#include <Rembedded.h>
#include <Rinternals.h>

double r_times(double a, double b) {
  SEXP fun;
  SEXP e;

  PROTECT(e = allocVector(LANGSXP, 3));
  fun = findFun(install("times"), R_GlobalEnv);
  if (fun == R_NilValue) {
    fprintf(stderr, "No definition for function times.\n");
    UNPROTECT(1);
    return (-1);
  }

  SETCAR(e, fun);

  SETCADR(e, ScalarReal(a));
  SETCADDR(e, ScalarReal(b));

  int error;
  SEXP ret = R_tryEval(e, R_GlobalEnv, &error);
  if (error) {
    return (-1);
  }

  UNPROTECT(1);
  return *REAL(ret);
}
