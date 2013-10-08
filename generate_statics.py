#!/usr/bin/env python
# coding: utf-8

"""
Simple script to generate sample statics for VoD and with ordering.
All data will be stored in ./static directory
"""

import json
import random

MAX = 10000
CATEGORIES = [2000, 3000, 4000, 5000, 6000]


def store_vod(i):
    """Prepare dict for VoD object, and store it in file."""
    json.dump(
        {
            'id': i,
            'title': 'title%s' % i,
            'categories': random.sample(
                CATEGORIES,
                random.randint(1, 2))
        },
        open('./static/vods%s.json' % i, 'w')
    )


def main():
    """Entry point :) Generates VoD and sorting files"""
    # pylint: disable=W0141
    map(store_vod, range(MAX))  # Screw you pylint,
                                # it looks better than list comprehension
    sorting = {}
    for i in range(10):
        sorting[i] = random.sample(range(MAX), random.randint(1000, 10000))
        random.shuffle(sorting[i])
    json.dump(sorting, open('./static/sorting.json', 'w'))


if __name__ == '__main__':
    main()
