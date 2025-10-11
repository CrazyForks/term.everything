PLUGIN_NAME := term_everything.so
C_INTEROP_LIB := ../c_interop/build/libinterop.a
C_INTEROP_DIR := ../c_interop

all: $(PLUGIN_NAME)

GO_SRCS := $(wildcard *.go)

$(PLUGIN_NAME): $(C_INTEROP_LIB) $(GO_SRCS)
	go build -buildmode=plugin -o $(PLUGIN_NAME) .

$(C_INTEROP_LIB):  $(wildcard $(C_INTEROP_DIR)/include/*.h) $(wildcard $(C_INTEROP_DIR)/src/*.cpp)
	$(MAKE) -C $(C_INTEROP_DIR)

clean:
	rm -f $(PLUGIN_NAME)
	$(MAKE) -C $(C_INTEROP_DIR) clean

.PHONY: all clean