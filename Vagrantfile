Vagrant.configure("2") do |config|
  config.vm.box = "ubuntu/bionic64"
  config.vm.network "private_network", ip: "192.168.33.10"
  config.vm.provider "virtualbox" do |vb|
    vb.memory = "9216"
    vb.cpus = 4
  end
  config.vm.provision "shell", inline: <<-SHELL
    apt-get update
    ssh-keygen -b 2048 -t rsa -f ~/.ssh/id_rsa -q -N ""
    cat ~/.ssh/id_rsa.pub >> ~/.ssh/authorized_keys
  SHELL
end
