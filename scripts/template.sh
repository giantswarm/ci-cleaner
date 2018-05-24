#!/bin/bash
TYPE=${1:-guest}

sed "s/{{type}}/${TYPE}/" template.yaml.tpl | tee template.yaml
