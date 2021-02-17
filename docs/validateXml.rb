#!/usr/bin/env ruby
require 'nokogiri'

xsd = Nokogiri::XML::Schema(File.read('DSpace_controlledvocabulary.xsd'))

xsd.validate('/Users/paulwalk/Dropbox/Reference/_ou/COAR/COAR_Vocabularies/skos_website_builder_git/webroot/static/resource_types.xml').each do |error|
  puts "#{error.line} :: #{error.message}"
end