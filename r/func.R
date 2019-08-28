# This function is invoked by the C program r_test

add1 <- function(a, b) {
  cat("R received: ", a, "\n");
  cat("R received: ", b, "\n");
  return(a + b)
}
