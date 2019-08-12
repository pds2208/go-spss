#ifndef _SAV_READER_H
#define _SAV_READER_H

struct Data* parse_sav(const char *input_file);

extern void goAddData(char *, char *);

struct Data {
    int var_count;

    char *data;
    unsigned long used;
    unsigned long have;

    char *header;
    unsigned long header_used;
    unsigned long header_have;

    char *buffer;
    unsigned long buffer_size;
};


#endif
