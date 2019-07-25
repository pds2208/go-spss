#ifndef _SAV_READER_H
#define _SAV_READER_H

int parse_sav(const char *input_file);

extern void goAddLine(char *);
extern void goAddHeaderLine(int , char *, int, int);

struct Sav {
    int var_count;
    char *data;
    unsigned long used;
    unsigned long have;
    char *buffer;
    unsigned long buffer_size;
};

#endif
