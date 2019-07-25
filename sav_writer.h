//
// Created by Paul Soule on 2019-07-25.
//

#ifndef READ_SAV_SAV_WRITER_H
#define READ_SAV_SAV_WRITER_H


typedef struct {
    int sav_type;
    const char *name;
    const char *label;
} SavHeader;

typedef struct {
    int sav_type;
    void *value;
} SavData;


int save_sav(const char *output_file, const char *label,
             SavHeader **sav_header, int column_cnt, int data_rows, SavData **sav_data);

#endif