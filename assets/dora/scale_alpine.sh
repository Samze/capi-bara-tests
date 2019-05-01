#!/bin/bash


function deploy_doras() {
for i in {1..20}
do  
    echo "Kicking off dora ${i}"
    (cf v3-zdt-push dora-alpine-${i} -o cfcapidocker/dora:alpine &)
done
}

function scale_doras() {
    for i in {1..20}
    do  
        echo "Scale doras ${i}"
        (cf v3-scale dora-alpine-${i} -i 2 -f &)
    done
}

function delete_doras() {
for i in {1..50}
do  
    echo "deleting dora ${i}"
    (cf delete dora-alpine-${i} -f &)
done
}


function main() {
    #deploy_doras
    scale_doras
    # delete_doras
}

main
