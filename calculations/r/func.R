# This function is invoked by the C program r_test

add1 <- function(a, n) {
  cat("R received: ", a, "\n");

  return(a + n)
}
