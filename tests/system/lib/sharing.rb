class Sharing
  attr_accessor :couch_id, :couch_rev, :rules, :members
  attr_reader :description, :app_slug

  def self.make_xor_key
    random = Random.new.bytes(8)
    res = []
    random.each_byte do |c|
      res << (c & 0xf)
      res << (c >> 4)
    end
    res
  end

  def self.xor_id(id, key)
    l = key.length
    buf = id.bytes.to_a
    buf.each_with_index do |c, i|
      if 48 <= c && c <= 57
        c = (c - 48) ^ key[i%l].ord
      elsif 97 <= c && c <= 102
        c = (c -  87) ^ key[i%l].ord
      elsif 65 <= c && c <= 70
        c = (c - 55) ^ key[i%l].ord
      else
        next
      end
      if c < 10
        buf[i] = c + 48
      else
        buf[i] = (c - 10) + 97
      end
    end
    buf.pack('c*')
  end

  def self.get_sharing_info(inst, sharing_id, doctype)
    opts = {
      accept: "application/vnd.api+json",
      content_type: "application/vnd.api+json",
      authorization: "Bearer #{inst.token_for doctype}"
    }
    res = inst.client["/sharings/#{sharing_id}"].get opts
    JSON.parse(res.body)["data"]
  end

  def self.get_shared_docs(inst, sharing_id, doctype)
    j = get_sharing_info inst, sharing_id, doctype
    j.dig "relationships", "shared_docs", "data"
  end

  def add_members(inst, contacts, doctype)
    opts = {
      accept: "application/vnd.api+json",
      content_type: "application/vnd.api+json",
      authorization: "Bearer #{inst.token_for doctype}"
    }
    data = {
      data: {
        relationships: {
          recipients: {
            data: contacts.map(&:as_reference)
          }
        }
      }
    }
    body = JSON.generate data
    res = inst.client["/sharings/#{@couch_id}/recipients"].post body, opts
    res.code
  end

  def read_only(inst, index)
    opts = {
      authorization: "Bearer #{inst.token_for Folder.doctype}"
    }
    res = inst.client["/sharings/#{@couch_id}/recipients/#{index}/readonly"].post "", opts
    res.code
  end

  def read_write(inst, index)
    opts = {
      authorization: "Bearer #{inst.token_for Folder.doctype}"
    }
    res = inst.client["/sharings/#{@couch_id}/recipients/#{index}/readonly"].delete opts
    res.code
  end

  def revoke_by_sharer(inst, doctype)
    opts = {
      authorization: "Bearer #{inst.token_for doctype}"
    }
    res = inst.client["/sharings/#{@couch_id}/recipients"].delete opts
    res.code
  end

  def revoke_recipient_by_sharer(inst, doctype, index)
    opts = {
      authorization: "Bearer #{inst.token_for doctype}"
    }
    res = inst.client["/sharings/#{@couch_id}/recipients/#{index}"].delete opts
    res.code
  end

  def revoke_recipient_by_itself(inst, doctype)
    opts = {
      authorization: "Bearer #{inst.token_for doctype}"
    }
    res = inst.client["/sharings/#{@couch_id}/recipients/self"].delete opts
    res.code
  end

  def initialize(opts = {})
    @description = opts[:description] || Faker::TvShows::DrWho.catch_phrase
    @app_slug = opts[:app_slug] || ""
    @rules = []
    @members = [] # Owner's instance + recipients contacts
  end

  def self.doctype
    "io.cozy.sharings"
  end

  def as_json_api
    recipients = @members.drop 1
    {
      data: {
        doctype: self.class.doctype,
        attributes: {
          description: @description,
          app_slug: @app_slug,
          rules: @rules.map(&:as_json),
          open_sharing: true
        },
        relationships: {
          recipients: {
            data: recipients.map(&:as_reference)
          }
        }
      }
    }
  end

  def owner
    @members.first
  end
end
