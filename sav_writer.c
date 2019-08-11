#include "sav_writer.h"
#include <fcntl.h>
#include <stdio.h>
#include <stdlib.h>
#include <unistd.h>

const int MAX_STRING = 4096;

readstat_variable_t *save_header(file_header *const *sav_header, int column_cnt,
                                 readstat_writer_t *writer);

static ssize_t write_bytes(const void *data, size_t len, void *ctx) {
    int fd = *(int *) ctx;
    return write(fd, data, len);
}

int save_sav(const char *output_file, const char *label, file_header **sav_header, int column_cnt,
             int data_rows, data_item **sav_data) {
    readstat_writer_t *writer = readstat_writer_init();
    readstat_set_data_writer(writer, &write_bytes);
    readstat_writer_set_file_label(writer, label);

    for (int i = 0; i < column_cnt; i++) {
        unsigned long cnt = 0;
        if (sav_header[i]->sav_type == READSTAT_TYPE_STRING) {
            cnt = MAX_STRING;
        }
        if (sav_header[i]->sav_type == READSTAT_TYPE_DOUBLE) {
            cnt = 8;
        }
        readstat_variable_t *variable =
                readstat_add_variable(writer, sav_header[i]->name, sav_header[i]->sav_type, cnt);
        sav_header[i]->variable = variable;
        readstat_variable_set_label(variable, sav_header[i]->label);
    }

    int fd = open(output_file, O_WRONLY | O_CREAT | O_TRUNC, 0666);

    if (fd == -1) {
        return -1;
    }
    readstat_begin_writing_sav(writer, &fd, data_rows);

    int cnt = 0;

    for (int i = 0; i < data_rows; i++) {
        readstat_begin_row(writer);

        for (int j = 0; j < column_cnt; j++) {
            readstat_variable_t *variable = sav_header[j]->variable;
            switch (sav_data[cnt]->sav_type) {
                case READSTAT_TYPE_STRING:
                    readstat_insert_string_value(writer, variable, (const char *) sav_data[cnt]->string_value);
                    break;

                case READSTAT_TYPE_INT8:
                    readstat_insert_int8_value(writer, variable, sav_data[cnt]->int_value);
                    break;

                case READSTAT_TYPE_INT16:
                    readstat_insert_int16_value(writer, variable, sav_data[cnt]->int_value);
                    break;

                case READSTAT_TYPE_INT32:
                    readstat_insert_int32_value(writer, variable, sav_data[cnt]->int_value);
                    break;

                case READSTAT_TYPE_FLOAT:
                    readstat_insert_float_value(writer, variable, sav_data[cnt]->float_value);
                    break;

                case READSTAT_TYPE_DOUBLE:
                    readstat_insert_double_value(writer, variable, sav_data[cnt]->double_value);
                    break;

                default:
                    break;
            }
            cnt++;
        }

        readstat_end_row(writer);
    }

    readstat_end_writing(writer);
    readstat_writer_free(writer);
    close(fd);

    return 0;
}
