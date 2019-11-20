#!/usr/bin/env bash

# ZUNKAPATH must be defined.
[[ -z "$ZUNKAPATH" ]] && printf "error: ZUNKAPATH enviorment not defined.\n" >&2 && exit 1 

# Create dir if not exist.
mkdir -p $ZUNKAPATH/xml/aldo
mkdir -p $ZUNKAPATH/log/aldo

# Last downloaded XML file.
F_LAST=$ZUNKAPATH/xml/aldo/aldo-products.xml
# XML file backup.
F_BACKUP=$ZUNKAPATH/xml/aldo/aldo-products-$(date +%Y-%h-%d-%H%M%S).xml
# Log file.
F_LOG=$ZUNKAPATH/log/aldo/aldo-xml-products.log

# Get xml file.
echo Downloading XML file.
# RESULT=`curl "http://webservice.aldo.com.br/asp.net/ferramentas/integracao.ashx?u=146612&p=zunk4c" \
# RESULT=`curl "https://www.zunka.com.br/xml/zoom/zoom-produtos.xml" \
RESULT=`curl "http://webservice.aldo.com.br/asp.net/ferramentas/integracao.ashx?u=146612&p=zunk4c" \
    -s -w "%{time_total} %{size_download}" \
    -o $F_BACKUP` 

TIME=`echo $RESULT | cut -d " " -f1`
SIZE=`echo $RESULT | cut -d " " -f2`
# Kilobytes.
SIZE=`expr $SIZE / 1024`
# Megabytes.
# SIZE=`expr $SIZE / 1048576`

# Log time and size.
printf "$(date +%FT%T%:z)  %.1fs  %.0fkb  $F_BACKUP\n" $TIME $SIZE >> $F_LOG

# Copy as last xml file.
cp $F_BACKUP $F_LAST

# Process xml file.
echo Processing XML file.
aldowsc < $F_LAST