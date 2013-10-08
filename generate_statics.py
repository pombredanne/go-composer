#!/usr/bin/env python
# coding: utf-8

import json
import random

MAX = 10000
CATEGORIES = [2000, 3000, 4000, 5000, 6000]


def main():
    map(lambda i: json.dump(
        {
            'id': i,
            'title': 'title%s' % i,
            'categories': random.sample(
                CATEGORIES,
                random.randint(1, 2))
        },
        open('./static/vods%s.json' % i, 'w')), range(MAX))
    sorting = {}
    for i in range(10):
        sorting[i] = random.sample(range(MAX), random.randint(1000, 10000))
        random.shuffle(sorting[i])
    json.dump(sorting, open('./static/sorting.json', 'w'))


if __name__ == '__main__':
    main()
