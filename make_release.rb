#!/usr/bin/env ruby

version = (File.read './VERSION').strip

['linux'].each do |os|
  ['amd64', 'arm64'].each do |arch|
    output_dir_relative = "db-#{version}-#{os}-#{arch}"
    output_dir = "./ignore/#{output_dir_relative}"
    output_tarball = "./ignore/#{output_dir_relative}.tar.gz"
    `mkdir -p #{output_dir}`

    # Build
    `GOOS=#{os} GOARCH=#{arch} go build -o #{output_dir}/db .`
    # Copy LICENSE file
    `cp ./LICENSE #{output_dir}`

    `tar cvzf #{output_tarball} -C ignore #{output_dir_relative}`
    `rm -r #{output_dir}`
  end
end
