#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include "readstat.h"

#include "sav_reader.h"

struct Sav {
    int var_count;
    char *data;
    unsigned long used;
    unsigned long have;
};

const int ACCOC_SIZE = 2 * 1024 * 1024;
const int BUFFER_SIZE = 4 * 1024;  // maximum size of a column's value in bytes.
// Bit over the top but meh

void add_to_buffer(const char *d, struct Sav *sav) {

    unsigned long len = strlen(d) + 1;

    if (sav->have < len) {
        sav->data = realloc(sav->data, sav->used + ACCOC_SIZE);
        sav->have += ACCOC_SIZE;
    }

    strcpy(&sav->data[sav->used], d);
    sav->used += len - 1;
    sav->have -= len;
}

int handle_metadata(readstat_metadata_t *metadata, void *ctx) {

    struct Sav *sav = (struct Sav *) ctx;
    sav->var_count = readstat_get_var_count(metadata);
    return READSTAT_HANDLER_OK;
}

int handle_variable(int index, readstat_variable_t *variable,
                    const char *val_labels, void *ctx) {

    struct Sav *sav = (struct Sav *) ctx;

    int var_index = readstat_variable_get_index(variable);
    readstat_type_t type = variable->type;

    const char *name = readstat_variable_get_name(variable);
    goAddHeaderLine(var_index, (char *) name, (int) type, 0);

    if (index == sav->var_count - 1) {
        goAddHeaderLine(var_index, (char *) name, (int) type, 1);
    }

    return READSTAT_HANDLER_OK;
}

int handle_value(int obs_index, readstat_variable_t *variable,
                 readstat_value_t value, void *ctx) {

    struct Sav *sav = (struct Sav *) ctx;
    int var_index = readstat_variable_get_index(variable);

    readstat_type_t type = readstat_value_type(value);
    const char *name = readstat_variable_get_name(variable);

    char buf[BUFFER_SIZE]; // this is quicker than malloc / free

    switch (type) {
        case READSTAT_TYPE_STRING:
            if (readstat_value_is_system_missing(value)) {
                snprintf(buf, sizeof(buf), "\"\"");
            } else {
                snprintf(buf, sizeof(buf), "\"%s\"", readstat_string_value(value));
            }
            add_to_buffer(buf, sav);
            break;

        case READSTAT_TYPE_INT8:
            if (readstat_value_is_system_missing(value)) {
                snprintf(buf, sizeof(buf), "%d", 0);
            } else {
                snprintf(buf, sizeof(buf), "%d", readstat_int8_value(value));
            }
            add_to_buffer(buf, sav);
            break;

        case READSTAT_TYPE_INT16:
            if (readstat_value_is_system_missing(value)) {
                snprintf(buf, sizeof(buf), "%d", 0);
            } else {
                snprintf(buf, sizeof(buf), "%d", readstat_int16_value(value));
            }
            add_to_buffer(buf, sav);
            break;

        case READSTAT_TYPE_INT32:
            if (readstat_value_is_system_missing(value)) {
                snprintf(buf, sizeof(buf), "%d", 0);
            } else {
                snprintf(buf, sizeof(buf), "%d", readstat_int32_value(value));
            }
            add_to_buffer(buf, sav);
            break;

        case READSTAT_TYPE_FLOAT:
            if (readstat_value_is_system_missing(value)) {
                snprintf(buf, sizeof(buf), "%f", 0.0);
            } else {
                snprintf(buf, sizeof(buf), "%f", readstat_float_value(value));
            }
            add_to_buffer(buf, sav);

            break;

        case READSTAT_TYPE_DOUBLE:
            if (readstat_value_is_system_missing(value)) {
                snprintf(buf, sizeof(buf), "%lf", 0.0);
            } else {
                snprintf(buf, sizeof(buf), "%lf", readstat_double_value(value));
            }
            add_to_buffer(buf, sav);

            break;

        default:
            return READSTAT_HANDLER_OK;
    }

    if (var_index == sav->var_count - 1) {
        goAddLine(sav->data);

        if (sav->data != NULL) {
            free(sav->data);
            sav->data = NULL;
            sav->used = 0;
            sav->have = 0;
        }
    } else {
        add_to_buffer(",", sav);
    }

    return READSTAT_HANDLER_OK;
}

int parse_sav(const char *input_file) {

    if (input_file == 0) {
        return -1;
    }

    readstat_error_t error;
    readstat_parser_t *parser = readstat_parser_init();
    readstat_set_metadata_handler(parser, &handle_metadata);
    readstat_set_variable_handler(parser, &handle_variable);
    readstat_set_value_handler(parser, &handle_value);

    struct Sav *sav = (struct Sav *) malloc(sizeof (struct Sav));

    sav->data = NULL;
    sav->used = 0;
    sav->have = 0;

    error = readstat_parse_sav(parser, input_file, sav);

    readstat_parser_free(parser);

    if (sav->data != NULL) {
        free(sav->data);
        free(sav);
    }

    if (error != READSTAT_OK) {
        return -1;
    }

    return 0;
}
