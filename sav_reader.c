#include "readstat.h"
#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "sav_reader.h"

int handle_metadata(readstat_metadata_t *metadata, void *ctx) {

    int *my_var_count = (int *) ctx;
    *my_var_count = readstat_get_var_count(metadata);
    return READSTAT_HANDLER_OK;
}

char *data = NULL;
int used = 0;
int have = 0;

const int ACCOC_SIZE = 2 * 1024 * 1024;
const int BUFFER_SIZE = 4 * 1024; // maximum size of a column's value in bytes. Bit over the top but meh

void add_to_buffer(const char *d) {
    int len = strlen(d) + 1;

    if (have < len) {
        data = realloc(data, used + ACCOC_SIZE);
        have += ACCOC_SIZE;
    }

    strcpy(&data[used], d);
    used += len - 1;
    have -= len;
}

int handle_variable(int index, readstat_variable_t *variable,
                    const char *val_labels, void *ctx) {

    int *my_var_count = (int *) ctx;
    int var_index = readstat_variable_get_index(variable);
    readstat_type_t type = variable->type;

    const char *name = readstat_variable_get_name(variable);
    goAddHeaderLine(var_index, (char *) name, (int) type, 0);

    if (index == *my_var_count - 1) {
        goAddHeaderLine(var_index, (char *) name, (int) type, 1);
    }

    return READSTAT_HANDLER_OK;
}

int handle_value(int obs_index, readstat_variable_t *variable,
                 readstat_value_t value, void *ctx) {

    int *my_var_count = (int *) ctx;
    int var_index = readstat_variable_get_index(variable);

    readstat_type_t type = readstat_value_type(value);
    const char *name = readstat_variable_get_name(variable);

    char buf[BUFFER_SIZE];

    switch (type) {

        case READSTAT_TYPE_STRING:
            if (readstat_value_is_system_missing(value)) {
                snprintf(buf, sizeof(buf), "\"\"");
            } else {
                snprintf(buf, sizeof(buf), "\"%s\"", readstat_string_value(value));
            }
            add_to_buffer(buf);
            break;

        case READSTAT_TYPE_INT8:
            if (readstat_value_is_system_missing(value)) {
                snprintf(buf, sizeof(buf), "0");
            } else {
                snprintf(buf, sizeof(buf), "%hhd", readstat_int8_value(value));
            }
            add_to_buffer(buf);
            break;

        case READSTAT_TYPE_INT16:
            if (readstat_value_is_system_missing(value)) {
                snprintf(buf, sizeof(buf), "0");
            } else {
                snprintf(buf, sizeof(buf), "%d", readstat_int16_value(value));
            }
            add_to_buffer(buf);
            break;

        case READSTAT_TYPE_INT32:
            if (readstat_value_is_system_missing(value)) {
                snprintf(buf, sizeof(buf), "0");
            } else {
                snprintf(buf, sizeof(buf), "%d", readstat_int32_value(value));
            }
            add_to_buffer(buf);
            break;

        case READSTAT_TYPE_FLOAT:
            if (readstat_value_is_system_missing(value)) {
                snprintf(buf, sizeof(buf), "0.0");
            } else {
                snprintf(buf, sizeof(buf), "%f", readstat_float_value(value));
            }
            add_to_buffer(buf);

            break;

        case READSTAT_TYPE_DOUBLE:
            if (readstat_value_is_system_missing(value)) {
                snprintf(buf, sizeof(buf), "0.0");
            } else {
                snprintf(buf, sizeof(buf), "%lf", readstat_double_value(value));
            }
            add_to_buffer(buf);

            break;

        default:
            return READSTAT_HANDLER_OK;
    }

    if (var_index == *my_var_count - 1) {
        goAddLine(data);

        if (data != NULL) {
            free(data);
            data = NULL;
            used = 0;
            have = 0;
        }
    } else {
        add_to_buffer(",");
    }

    return READSTAT_HANDLER_OK;
}

int parse_sav(const char *input_file) {

    if (input_file == 0) {
        return -1;
    }

    int my_var_count = 0;
    readstat_error_t error = READSTAT_OK;
    readstat_parser_t *parser = readstat_parser_init();
    readstat_set_metadata_handler(parser, &handle_metadata);
    readstat_set_variable_handler(parser, &handle_variable);
    readstat_set_value_handler(parser, &handle_value);

    error = readstat_parse_sav(parser, input_file, &my_var_count);
    readstat_parser_free(parser);

    if (data != NULL) {
        free(data);
    }

    if (error != READSTAT_OK) {
        return -1;
    }

    return 0;
}
