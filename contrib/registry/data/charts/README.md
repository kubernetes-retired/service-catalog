# Charts

## Releasing charts

Our charts are hosted in GCS in the bucket `helm-sb-test`. Follow these steps
to release a new version of a chart.

To release new versions of the Helm charts, update the chart, **and** its
version number and then run:

    ./update.sh

Afterwards, update the version number in the `definitions.json` file for the
registry to pick it up, and commit the change.
