// Code generated by sqlc-grpc (https://github.com/walterwanderley/sqlc-grpc). DO NOT EDIT.

package books

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	pb "booktest/api/books/v1"
	"booktest/internal/validation"
)

type Service struct {
	pb.UnimplementedBooksServiceServer
	logger  *zap.Logger
	querier *Queries
}

func (s *Service) BooksByTags(ctx context.Context, in *pb.BooksByTagsRequest) (out *pb.BooksByTagsResponse, err error) {
	dollar_1 := in.GetDollar_1()

	result, err := s.querier.BooksByTags(ctx, dollar_1)
	if err != nil {
		s.logger.Error("BooksByTags sql call failed", zap.Error(err))
		return
	}
	out = new(pb.BooksByTagsResponse)
	for _, r := range result {
		var item *pb.BooksByTagsRow
		item, err = toBooksByTagsRow(r)
		if err != nil {
			return
		}
		out.Value = append(out.Value, item)
	}
	return
}

func (s *Service) BooksByTitleYear(ctx context.Context, in *pb.BooksByTitleYearRequest) (out *pb.BooksByTitleYearResponse, err error) {
	var arg BooksByTitleYearParams
	arg.Title = in.GetTitle()
	arg.Year = in.GetYear()

	result, err := s.querier.BooksByTitleYear(ctx, arg)
	if err != nil {
		s.logger.Error("BooksByTitleYear sql call failed", zap.Error(err))
		return
	}
	out = new(pb.BooksByTitleYearResponse)
	for _, r := range result {
		var item *pb.Book
		item, err = toBook(r)
		if err != nil {
			return
		}
		out.Value = append(out.Value, item)
	}
	return
}

func (s *Service) CreateAuthor(ctx context.Context, in *pb.CreateAuthorRequest) (out *pb.Author, err error) {
	name := in.GetName()

	result, err := s.querier.CreateAuthor(ctx, name)
	if err != nil {
		s.logger.Error("CreateAuthor sql call failed", zap.Error(err))
		return
	}
	return toAuthor(result)
}

func (s *Service) CreateBook(ctx context.Context, in *pb.CreateBookRequest) (out *pb.Book, err error) {
	var arg CreateBookParams
	arg.AuthorID = in.GetAuthorId()
	arg.Isbn = in.GetIsbn()
	arg.BookType = BookType(in.GetBookType())
	arg.Title = in.GetTitle()
	arg.Year = in.GetYear()
	if v := in.GetAvailable(); v != nil {
		if err = v.CheckValid(); err != nil {
			err = fmt.Errorf("invalid Available: %s%w", err.Error(), validation.ErrUserInput)
			return
		}
		arg.Available = v.AsTime()
	} else {
		err = fmt.Errorf("field Available is required%w", validation.ErrUserInput)
		return
	}
	arg.Tags = in.GetTags()

	result, err := s.querier.CreateBook(ctx, arg)
	if err != nil {
		s.logger.Error("CreateBook sql call failed", zap.Error(err))
		return
	}
	return toBook(result)
}

func (s *Service) DeleteBook(ctx context.Context, in *pb.DeleteBookRequest) (out *pb.DeleteBookResponse, err error) {
	bookID := in.GetBookId()

	err = s.querier.DeleteBook(ctx, bookID)
	if err != nil {
		s.logger.Error("DeleteBook sql call failed", zap.Error(err))
		return
	}
	out = new(pb.DeleteBookResponse)
	return
}

func (s *Service) GetAuthor(ctx context.Context, in *pb.GetAuthorRequest) (out *pb.Author, err error) {
	authorID := in.GetAuthorId()

	result, err := s.querier.GetAuthor(ctx, authorID)
	if err != nil {
		s.logger.Error("GetAuthor sql call failed", zap.Error(err))
		return
	}
	return toAuthor(result)
}

func (s *Service) GetBook(ctx context.Context, in *pb.GetBookRequest) (out *pb.Book, err error) {
	bookID := in.GetBookId()

	result, err := s.querier.GetBook(ctx, bookID)
	if err != nil {
		s.logger.Error("GetBook sql call failed", zap.Error(err))
		return
	}
	return toBook(result)
}

func (s *Service) UpdateBook(ctx context.Context, in *pb.UpdateBookRequest) (out *pb.UpdateBookResponse, err error) {
	var arg UpdateBookParams
	arg.Title = in.GetTitle()
	arg.Tags = in.GetTags()
	arg.BookType = BookType(in.GetBookType())
	arg.BookID = in.GetBookId()

	err = s.querier.UpdateBook(ctx, arg)
	if err != nil {
		s.logger.Error("UpdateBook sql call failed", zap.Error(err))
		return
	}
	out = new(pb.UpdateBookResponse)
	return
}

func (s *Service) UpdateBookISBN(ctx context.Context, in *pb.UpdateBookISBNRequest) (out *pb.UpdateBookISBNResponse, err error) {
	var arg UpdateBookISBNParams
	arg.Title = in.GetTitle()
	arg.Tags = in.GetTags()
	arg.BookID = in.GetBookId()
	arg.Isbn = in.GetIsbn()

	err = s.querier.UpdateBookISBN(ctx, arg)
	if err != nil {
		s.logger.Error("UpdateBookISBN sql call failed", zap.Error(err))
		return
	}
	out = new(pb.UpdateBookISBNResponse)
	return
}
