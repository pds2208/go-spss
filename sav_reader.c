#include <stdint.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

#include "readstat.h"
#include "sav_reader.h"

const int ACCOC_SIZE = 256 * 1024 * 1024;
const int SAV_BUFFER_SIZE = 128;  // initial buffer size for a value, will grow if necessary

void add_to_data(const char *d, struct Data *data) {
    unsigned long len = strlen(d) + 1;

    if (data->have < len) {
        data->data = realloc(data->data, data->used + ACCOC_SIZE);
        data->have += ACCOC_SIZE;
    }

    strcpy(&data->data[data->used], d);
    data->used += len - 1;
    data->have -= len;
}

void add_to_header(const char *d, struct Data *data) {
    unsigned long len = strlen(d) + 1;

    if (data->header_have < len) {
        data->header = realloc(data->header, data->header_used + ACCOC_SIZE);
        data->header_have += ACCOC_SIZE;
    }

    strcpy(&data->header[data->header_used], d);
    data->header_used += len - 1;
    data->header_have -= len;
}

int handle_metadata(readstat_metadata_t *metadata, void *ctx) {
    struct Data *sav = (struct Data *) ctx;
    sav->var_count = readstat_get_var_count(metadata);
    return READSTAT_HANDLER_OK;
}

int handle_variable(int index, readstat_variable_t *variable, const char *val_labels, void *ctx) {
    struct Data *sav = (struct Data *) ctx;
    const char *name = readstat_variable_get_name(variable);
    int var_index = readstat_variable_get_index(variable);

    if (var_index == sav->var_count - 1) {
       add_to_header(name, sav);
       add_to_header("\n", sav);
    } else {
       add_to_header(name, sav);
       add_to_header(",", sav);
    }

    return READSTAT_HANDLER_OK;
}

int handle_value(int obs_index, readstat_variable_t *variable, readstat_value_t value, void *ctx) {
    struct Data *sav = (struct Data *) ctx;
    int var_index = readstat_variable_get_index(variable);

    readstat_type_t type = readstat_value_type(value);

    char *buf = sav->buffer;

    switch (type) {
        case READSTAT_TYPE_STRING:

            // This will be the only place we can expect a value larger than the
            // existing SAV_BUFFER_SIZE
            // We use snprintf as it's much faster
            if (sav->buffer_size <= strlen(readstat_string_value(value)) + 1) {
                sav->buffer_size = strlen(readstat_string_value(value)) + SAV_BUFFER_SIZE + 1;
                sav->buffer = realloc(sav->buffer, sav->buffer_size);
            }
            char *str = (char *) readstat_string_value(value);
            for (char* p = str; (p = strchr(p, ',')) ; ++p) {
                *p = ' ';
            }
            snprintf(buf, sav->buffer_size, "%s", readstat_string_value(value));

            add_to_data(buf, sav);
            break;

        case READSTAT_TYPE_INT8:
            if (readstat_value_is_system_missing(value)) {
                snprintf(buf, sav->buffer_size, "%d", 0);
            } else {
                snprintf(buf, sav->buffer_size, "%d", readstat_int8_value(value));
            }
            add_to_data(buf, sav);
            break;

        case READSTAT_TYPE_INT16:
            if (readstat_value_is_system_missing(value)) {
                snprintf(buf, sav->buffer_size, "%d", 0);
            } else {
                snprintf(buf, sav->buffer_size, "%d", readstat_int16_value(value));
            }
            add_to_data(buf, sav);
            break;

        case READSTAT_TYPE_INT32:
            if (readstat_value_is_system_missing(value)) {
                snprintf(buf, sav->buffer_size, "%d", 0);
            } else {
                snprintf(buf, sav->buffer_size, "%d", readstat_int32_value(value));
            }
            add_to_data(buf, sav);
            break;

        case READSTAT_TYPE_FLOAT:
            if (readstat_value_is_system_missing(value)) {
                snprintf(buf, sav->buffer_size, "%f", 0.0);
            } else {
                snprintf(buf, sav->buffer_size, "%f", readstat_float_value(value));
            }
            add_to_data(buf, sav);

            break;

        case READSTAT_TYPE_DOUBLE:
            if (readstat_value_is_system_missing(value)) {
                snprintf(buf, sav->buffer_size, "%lf", 0.0);
            } else {
                snprintf(buf, sav->buffer_size, "%lf", readstat_double_value(value));
            }
            add_to_data(buf, sav);

            break;

        default:
            return READSTAT_HANDLER_OK;
    }

    if (var_index == sav->var_count - 1) {
        add_to_data("\n", sav);
    } else {
        add_to_data(",", sav);
    }

    return READSTAT_HANDLER_OK;
}

struct Data * parse_sav(const char *input_file) {

    if (input_file == 0) {
        return NULL;
    }

    readstat_error_t error;
    readstat_parser_t *parser = readstat_parser_init();
    readstat_set_metadata_handler(parser, &handle_metadata);
    readstat_set_variable_handler(parser, &handle_variable);
    readstat_set_value_handler(parser, &handle_value);

    struct Data *sav_data = (struct Data *) malloc(sizeof(struct Data));
    sav_data->data = NULL;
    sav_data->used = 0;
    sav_data->have = 0;

    sav_data->buffer = malloc(SAV_BUFFER_SIZE);
    sav_data->buffer_size = SAV_BUFFER_SIZE;

    sav_data->header = NULL;
    sav_data->header_have = 0;
    sav_data->header_used = 0;

    error = readstat_parse_sav(parser, input_file, sav_data);

    readstat_parser_free(parser);

    if (error != READSTAT_OK) {
      return NULL;
    }

    // remove the final newline character
    sav_data->data[sav_data->used -1] = 0;

    return sav_data;

}
