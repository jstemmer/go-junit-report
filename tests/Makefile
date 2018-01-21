CC = g++
LIBS =
CFLAGS = -g -Wall -MMD -MP
LDFLAGS =

# output files
BIN = bin_test

# directories
INC_DIR = ../../doctest

# files
SRCS = $(filter-out ./main.cpp, $(shell find . -name "*.cpp"))
OBJS = $(SRCS:.cpp=.o)
BINS = $(SRCS:.cpp=)
DEPS = $(OBJS:.o=.d)

# tasks
all: $(BINS)

$(BINS): main.o $(OBJS)
	@for obj in $(OBJS) ; do \
		echo "link "$${obj%.o}"" ; \
		$(CC) -o "$${obj%.o}" main.o "$$obj" $(LDFLAGS) $(LIBS) ; \
	done

%.o: %.cpp
	$(CC) $(CFLAGS) -I$(INC_DIR) -o $@ -c $<

exec: $(BINS)
	@for bin in $(BINS) ; do \
		$${bin} -d -s > $${bin}.txt || echo "Finished with fail: "$${bin}"" ; \
	done

CLEAN_TARGETS = $(BINS) *.o *.d *.txt
clean:
	rm -rf $(CLEAN_TARGETS)

.PHONY: all clean

-include $(DEPS)
