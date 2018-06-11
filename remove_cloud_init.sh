#!/bin/bash

echo 'datasource_list: [ None ]' | sudo -s tee /etc/cloud/cloud.cfg.d/90_dpkg.cfg
sudo apt-get purge -y cloud-init
sudo rm -rf /etc/cloud/; sudo rm -rf /var/lib/cloud/

echo "Cloud-init removed"
