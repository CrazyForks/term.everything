.PHONY: all clean build

.DELETE_ON_ERROR:


bin_name := term.everything❗mmulet.com-dont_forget_to_chmod_+x_this_file

protocols_files := $(shell find ./wayland/generate)

xml_protocols := $(shell find ./wayland/generate -name "*.xml")

generated_protocols := $(patsubst ./wayland/generate/resources/%,./wayland/protocols/%.go,$(xml_protocols))

generated_helpers := $(patsubst ./wayland/generate/resources/%,./wayland/%.helper.go,$(xml_protocols))

build: $(generated_protocols) $(generated_helpers) $(bin_name)

# grouped target to generate all protocols and helpers in one go
# the & is what does this
$(generated_protocols) $(generated_helpers)&: $(protocols_files) ./wayland/generate.go
	go generate ./wayland

$(bin_name): go.mod main.go $(shell find ./wayland) $(shell find ./termeverything) Makefile $(shell find ./framebuffertoansi) $(shell find ./escapecodes) $(generated_protocols) $(generated_helpers)
	go build -o $(bin_name) .

clean:
	rm __debug_bin* || true
	rm term.everything || true
	rm term.everything❗mmulet.com-dont_forget_to_chmod_+x_this_file || true
	rm ./wayland/protocols/*.xml.go || true
	rm ./wayland/*.helper.go || true