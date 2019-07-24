#ifndef _SAV_READER_H
#define _SAV_READER_H

int parse_sav(const char *input_file);

extern void goAddLine(char *);
extern void goAddHeaderLine(int , char *, int, int);

#endif
