# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant.configure("2") do |config|
  config.vm.box = "bento/ubuntu-16.04"

  config.vm.network "private_network", ip: "192.168.12.34"

  config.vm.provider "virtualbox" do |vb|
    vb.memory = "512"
    vb.cpus = "2"
  end
end