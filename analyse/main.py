#!/usr/bin/python

import sys
import logging
import psycopg2
import json
import numpy
import datetime
import argparse

logging.basicConfig()

logger = logging.getLogger()
logger.setLevel(logging.INFO)

class Analyzer:
    def __init__(self, pg):
        self.pg = pg

    def load(self, source_id):
        self.times = []
        self.ids = []
        self.data = {}
        with self.pg.cursor("fk_records") as pgc:
            pgc.execute("SELECT id, source_id, timestamp, data FROM fieldkit.record WHERE source_id = %s ORDER BY timestamp", (source_id,))
            self.all_keys = set([])
            prev_keys = set([])
            for record in pgc:
                source_id = record[1]
                timestamp = record[2]
                json = record[3]
                new_keys = set(json.keys())
                self.all_keys = self.all_keys.union(new_keys)
                if prev_keys != new_keys:
                    logger.info("New Keys: {} {} -> {}".format(timestamp, source_id, new_keys))
                    prev_keys = new_keys

                self.times.append(timestamp)
                self.ids.append(record[0])

                for key in self.all_keys:
                    if key in json:
                        self.data.setdefault(key, []).append(json[key])
                    else:
                        self.data.setdefault(key, []).append(None)

    def outliers(self):
        for key in self.all_keys:
            with_indices = [[i, v] for i, v in enumerate(self.data[key]) if v != None and v != 'NaN']

            if len(with_indices) == 0:
                logger.info("Processing {} ({} values) no data.".format(key, len(with_indices)))
                continue

            array = numpy.array(with_indices)
            column = array[:,1]
            median = numpy.median(column)
            mean = numpy.mean(column)
            sd = numpy.std(column)
            d = numpy.abs(column - numpy.median(column))
            mdev = numpy.median(d)
            s = d / mdev if mdev else None
            m = 3.5
            m = 60

            if s is None:
                logger.info("Processing {} ({} values) no deviation.".format(key, len(with_indices)))
                continue

            logger.info("Processing {} ({} values) mean = {} median = {} sd = {}".format(key, len(with_indices), mean, median, sd))

            excluded = [[self.ids[x[0]], x[1]] for i, x in enumerate(with_indices) if (s[i] > m)]

            if len(excluded) > 0:
                logger.info("Excluded {} values".format(len(excluded)))
                logger.info("Excluded: {}".format(excluded))
                ids = [row[0] for row in excluded]
                with self.pg.cursor() as pgc:
                    for id in ids:
                        pgc.execute("INSERT INTO fieldkit.record_analysis (record_id, outlier) VALUES (%s, %s) ON CONFLICT (record_id) DO UPDATE SET outlier = excluded.outlier", (id, True))

def main():
    logger.info("Starting...")

    parser = argparse.ArgumentParser(description='Analyse FK record data.')
    parser.add_argument("--postgres-url", default="postgres://fieldkit:password@127.0.0.1/fieldkit?sslmode=disable", action='store')
    args = parser.parse_args()

    with psycopg2.connect(args.postgres_url) as pg:
        for source_id in [4, 125, 126]:
            analyzer = Analyzer(pg)
            analyzer.load(source_id)
            analyzer.outliers()

    return True

main()
