source "https://rubygems.org"

git_source(:github) do |repo_name|

"https://github.com/#{repo_name}.git"

end

ruby "3.1.2"

gem "rails", "~> 7.0.4"

gem "pg", "~> 1.1"

gem "puma", "~> 5.0"

gem "sprockets-rails"

gem "tzinfo-data", platforms: [:mingw, :mswin, :x64_mingw, :jruby]

gem "bootsnap", require: false

group :development, :test do

gem "debug", platforms: [:mri, :mingw, :x64_mingw]

end

group :development do

gem "web-console"

gem "error_highlight"

end

group :test do

gem "capybara"

gem "selenium-webdriver"

gem "webdrivers"

end
