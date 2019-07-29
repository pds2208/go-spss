//
// Created by Paul Soule on 2019-07-25.
//

#ifndef READ_SAV_SAV_WRITER_H
#define READ_SAV_SAV_WRITER_H

#include <readstat.h>

typedef struct {
    int sav_type;
    const char *name;
    const char *label;
    readstat_variable_t *variable;
} SavHeader;

typedef struct {
    const int sav_type;

    const char *string_value;
    const int int_value;
    const float float_value;
    const double double_value;
} SavData;

int save_sav(const char *output_file, const char *label,
             SavHeader **sav_header, int column_cnt, int data_rows, SavData **sav_data);

#endif